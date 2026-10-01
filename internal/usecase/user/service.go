package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// Repository is what this usecase needs from storage.
// It is defined here (the consumer), and implemented in repository/postgres.
type Repository interface {
	Create(ctx context.Context, u domain.User) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}

// PasswordHasher is implemented in internal/security.
type PasswordHasher interface {
	Hash(plain string) (string, error)
	Compare(hash, plain string) error
}

type Service struct {
	repo   Repository
	hasher PasswordHasher
}

func NewService(repo Repository, hasher PasswordHasher) *Service {
	return &Service{repo: repo, hasher: hasher}
}

type RegisterInput struct {
	Email    string
	Phone    string
	Password string
	FullName string
}
type LoginInput struct {
	Email    string
	Password string
}

func (s *Service) Register(ctx context.Context, in RegisterInput) (domain.User, error) {
	hash, err := s.hasher.Hash(in.Password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}

	return s.repo.Create(ctx, domain.User{
		Email:        normalizeEmail(in.Email),
		Phone:        strings.TrimSpace(in.Phone),
		PasswordHash: hash,
		FullName:     strings.TrimSpace(in.FullName),
	})
}
func (s *Service) Login(ctx context.Context, in LoginInput) (domain.User, error) {
	u, err := s.repo.GetByEmail(ctx, normalizeEmail(in.Email))
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return domain.User{}, err
	}

	if err := s.hasher.Compare(u.PasswordHash, in.Password); err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	if !u.IsActive {
		return domain.User{}, domain.ErrUserInactive
	}

	if err := s.repo.UpdateLastLogin(ctx, u.ID); err != nil {
		return domain.User{}, err
	}
	return u, nil

}

// Authenticate checks credentials. Login (JWT) will call this later.
// Wrong email and wrong password return the same error on purpose.
func (s *Service) Authenticate(ctx context.Context, email, password string) (domain.User, error) {
	u, err := s.repo.GetByEmail(ctx, normalizeEmail(email))
	if errors.Is(err, domain.ErrNotFound) {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	if err != nil {
		return domain.User{}, err
	}

	if err := s.hasher.Compare(u.PasswordHash, password); err != nil {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	if !u.IsActive {
		return domain.User{}, domain.ErrUserInactive
	}

	if err := s.repo.UpdateLastLogin(ctx, u.ID); err != nil {
		return domain.User{}, err
	}
	return u, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return s.repo.GetByID(ctx, id)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
