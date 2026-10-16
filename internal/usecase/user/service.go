package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// Repository is what this usecase needs from the users table (implemented by postgres.UserRepo).
type Repository interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

// RefreshTokenRepository is what this usecase needs from the refresh_tokens table
// (implemented by postgres.RefreshTokenRepo).
type RefreshTokenRepository interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash, userAgent string, expiresAt time.Time) (domain.RefreshToken, error)
	GetByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error)
	Rotate(ctx context.Context, oldID, userID uuid.UUID, newHash, userAgent string, newExpiresAt time.Time) error
	RevokeByHash(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
}

// PasswordHasher is implemented in internal/security.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}

// TokenIssuer is implemented by *security.JWTIssuer. It creates access tokens (JWT)
// and creates/hashes refresh tokens (random strings; only the hash is stored).
type TokenIssuer interface {
	Issue(userID uuid.UUID, role domain.Role) (string, error)
	NewRefreshToken() (plain, hash string, err error)
	HashRefreshToken(plain string) string
}

// Config holds the two lifetimes. AccessTTL is only used to tell the client "expires_in".
type Config struct {
	AccessTTL  time.Duration // e.g. 15 minutes
	RefreshTTL time.Duration // e.g. 30 days
}

// AuthResult is the outcome of register, login and refresh.
type AuthResult struct {
	User         domain.User
	AccessToken  string
	RefreshToken string
	ExpiresIn    time.Duration // how long the ACCESS token lives
}

type Service struct {
	repo        Repository
	refreshRepo RefreshTokenRepository
	hasher      PasswordHasher
	tokens      TokenIssuer
	cfg         Config
}

func NewService(repo Repository, refreshRepo RefreshTokenRepository, hasher PasswordHasher, tokens TokenIssuer, cfg Config) *Service {
	return &Service{repo: repo, refreshRepo: refreshRepo, hasher: hasher, tokens: tokens, cfg: cfg}
}

type RegisterInput struct {
	Email     string
	Phone     string
	Password  string
	FullName  string
	UserAgent string
}

type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
}

type RefreshInput struct {
	RefreshToken string
	UserAgent    string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (AuthResult, error) {
	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return AuthResult{}, fmt.Errorf("hash password: %w", err)
	}

	u, err := s.repo.Create(ctx, domain.User{
		Email:        normalizeEmail(in.Email),
		Phone:        strings.TrimSpace(in.Phone),
		PasswordHash: hash,
		FullName:     strings.TrimSpace(in.FullName),
	})
	if err != nil {
		return AuthResult{}, err
	}
	return s.startSession(ctx, u, in.UserAgent)
}

// Login checks credentials and returns both tokens.
// Wrong email and wrong password return the same error on purpose.
func (s *Service) Login(ctx context.Context, in LoginInput) (AuthResult, error) {
	u, err := s.repo.GetByEmail(ctx, normalizeEmail(in.Email))
	if errors.Is(err, domain.ErrNotFound) {
		return AuthResult{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return AuthResult{}, err
	}

	if err := s.hasher.Compare(u.PasswordHash, in.Password); err != nil {
		return AuthResult{}, domain.ErrInvalidCredentials
	}
	if !u.IsActive {
		return AuthResult{}, domain.ErrUserInactive
	}

	if err := s.repo.UpdateLastLogin(ctx, u.ID); err != nil {
		return AuthResult{}, err
	}
	return s.startSession(ctx, u, in.UserAgent)
}

// Refresh exchanges a valid refresh token for a NEW pair (rotation).
// Every failure is domain.ErrInvalidRefreshToken (401): the client must log in again.
func (s *Service) Refresh(ctx context.Context, in RefreshInput) (AuthResult, error) {
	old, err := s.refreshRepo.GetByHash(ctx, s.tokens.HashRefreshToken(in.RefreshToken))
	if errors.Is(err, domain.ErrNotFound) {
		return AuthResult{}, domain.ErrInvalidRefreshToken // we never issued this token
	}
	if err != nil {
		return AuthResult{}, err
	}

	// A token that was already used is being presented AGAIN. Either a bug, or a thief with a copy.
	// Be safe: end every session of this user.
	if old.IsRevoked() {
		if err := s.refreshRepo.RevokeAllForUser(ctx, old.UserID); err != nil {
			return AuthResult{}, err
		}
		return AuthResult{}, domain.ErrInvalidRefreshToken
	}
	if old.IsExpired(time.Now()) {
		return AuthResult{}, domain.ErrInvalidRefreshToken
	}

	// Load the user NOW: the new access token gets the CURRENT role from the database,
	// and a disabled user cannot refresh.
	u, err := s.repo.GetByID(ctx, old.UserID)
	if errors.Is(err, domain.ErrNotFound) {
		return AuthResult{}, domain.ErrInvalidRefreshToken
	}
	if err != nil {
		return AuthResult{}, err
	}
	if !u.IsActive {
		return AuthResult{}, domain.ErrInvalidRefreshToken
	}

	access, err := s.tokens.Issue(u.ID, u.Role)
	if err != nil {
		return AuthResult{}, fmt.Errorf("issue access token: %w", err)
	}
	plain, newHash, err := s.tokens.NewRefreshToken()
	if err != nil {
		return AuthResult{}, fmt.Errorf("create refresh token: %w", err)
	}

	// Revoke the old and store the new in ONE transaction. If another request used the old token
	// first, Rotate returns ErrInvalidRefreshToken.
	err = s.refreshRepo.Rotate(ctx, old.ID, u.ID, newHash, clip(in.UserAgent, 255), time.Now().Add(s.cfg.RefreshTTL))
	if err != nil {
		return AuthResult{}, err
	}

	return AuthResult{User: u, AccessToken: access, RefreshToken: plain, ExpiresIn: s.cfg.AccessTTL}, nil
}

// Logout ends ONE session. An unknown token is not an error (logout is idempotent).
// The access token that was already issued stays valid until it expires: that is why it is short.
func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.refreshRepo.RevokeByHash(ctx, s.tokens.HashRefreshToken(refreshToken))
}

// LogoutAll ends every session of a user.
func (s *Service) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.refreshRepo.RevokeAllForUser(ctx, userID)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

// startSession is shared by Register and Login: it creates the two tokens and stores the refresh hash.
func (s *Service) startSession(ctx context.Context, u domain.User, userAgent string) (AuthResult, error) {
	access, err := s.tokens.Issue(u.ID, u.Role)
	if err != nil {
		return AuthResult{}, fmt.Errorf("issue access token: %w", err)
	}

	plain, hash, err := s.tokens.NewRefreshToken()
	if err != nil {
		return AuthResult{}, fmt.Errorf("create refresh token: %w", err)
	}
	if _, err := s.refreshRepo.Create(ctx, u.ID, hash, clip(userAgent, 255), time.Now().Add(s.cfg.RefreshTTL)); err != nil {
		return AuthResult{}, err
	}

	return AuthResult{User: u, AccessToken: access, RefreshToken: plain, ExpiresIn: s.cfg.AccessTTL}, nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// clip keeps a value inside its VARCHAR size (the User-Agent header can be long).
func clip(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}
