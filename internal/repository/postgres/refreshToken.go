package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
)

// RefreshTokenRepo needs the pool because Rotate uses a transaction.
// In routes.go: postgres.NewRefreshTokenRepo(a.DB)
type RefreshTokenRepo struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewRefreshTokenRepo(pool *pgxpool.Pool) *RefreshTokenRepo {
	return &RefreshTokenRepo{pool: pool, q: sqlc.New(pool)}
}

// Create stores the HASH of a new refresh token (the token itself is never stored).
func (r *RefreshTokenRepo) Create(ctx context.Context, userID uuid.UUID, tokenHash, userAgent string, expiresAt time.Time) (domain.RefreshToken, error) {
	row, err := r.q.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		UserAgent: userAgent,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return domain.RefreshToken{}, mapRefreshErr(err)
	}
	return toDomainRefreshToken(row), nil
}

// GetByHash finds a token record by the hash of what the client sent. No row -> ErrNotFound.
func (r *RefreshTokenRepo) GetByHash(ctx context.Context, tokenHash string) (domain.RefreshToken, error) {
	row, err := r.q.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return domain.RefreshToken{}, mapRefreshErr(err)
	}
	return toDomainRefreshToken(row), nil
}

// Rotate replaces an old token with a new one in ONE transaction:
// revoke the old, create the new.
// "0 rows revoked" means someone else used the old token a moment ago (two requests at once,
// or a replay), so it is refused: ErrInvalidRefreshToken. The transaction rolls back and nothing is created.
func (r *RefreshTokenRepo) Rotate(ctx context.Context, oldID, userID uuid.UUID, newHash, userAgent string, newExpiresAt time.Time) error {
	err := withTx(ctx, r.pool, func(q *sqlc.Queries) error {
		n, err := q.RevokeRefreshToken(ctx, oldID)
		if err != nil {
			return err
		}
		if n == 0 {
			return domain.ErrInvalidRefreshToken
		}

		_, err = q.CreateRefreshToken(ctx, sqlc.CreateRefreshTokenParams{
			UserID:    userID,
			TokenHash: newHash,
			UserAgent: userAgent,
			ExpiresAt: newExpiresAt,
		})
		return err
	})
	return mapRefreshErr(err)
}

// RevokeByHash is logout. An unknown or already revoked token is not an error: logout stays idempotent.
func (r *RefreshTokenRepo) RevokeByHash(ctx context.Context, tokenHash string) error {
	_, err := r.q.RevokeRefreshTokenByHash(ctx, tokenHash)
	return mapRefreshErr(err)
}

// RevokeAllForUser logs the user out everywhere.
func (r *RefreshTokenRepo) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	return mapRefreshErr(r.q.RevokeAllUserRefreshTokens(ctx, userID))
}

func toDomainRefreshToken(t sqlc.RefreshToken) domain.RefreshToken {
	return domain.RefreshToken{
		ID:        t.ID,
		UserID:    t.UserID,
		ExpiresAt: t.ExpiresAt,
		RevokedAt: t.RevokedAt,
		CreatedAt: t.CreatedAt,
	}
}

func mapRefreshErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, domain.ErrInvalidRefreshToken) {
		return err // already a domain error (from Rotate), pass it through
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign key: the user no longer exists
		return domain.ErrInvalidRefreshToken
	}
	return fmt.Errorf("refresh token repo: %w", err)
}
