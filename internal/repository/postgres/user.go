package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
)

const pgUniqueViolation = "23505"

// UserRepo is the only place that knows both sqlc and domain.
type UserRepo struct {
	q sqlc.Querier
}

func NewUserRepo(q sqlc.Querier) *UserRepo {
	return &UserRepo{q: q}
}

func (r *UserRepo) Create(ctx context.Context, u domain.User) (domain.User, error) {
	row, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
		Email:        u.Email,
		Phone:        u.Phone,
		PasswordHash: u.PasswordHash,
		FullName:     u.FullName,
	})
	if err != nil {
		return domain.User{}, mapUserErr(err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, mapUserErr(err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return domain.User{}, mapUserErr(err)
	}
	return toDomainUser(row), nil
}

func (r *UserRepo) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	if err := r.q.UpdateLastLogin(ctx, id); err != nil {
		return mapUserErr(err)
	}
	return nil
}

// toDomainUser converts the generated sqlc struct to the domain entity.
func toDomainUser(u sqlc.User) domain.User {
	return domain.User{
		ID:           u.ID,
		Email:        u.Email,
		Phone:        u.Phone,
		PasswordHash: u.PasswordHash,
		FullName:     u.FullName,
		Role:         domain.Role(u.Role),
		IsActive:     u.IsActive,
		LastLoginAt:  u.LastLoginAt,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// mapUserErr translates DB errors into domain errors.
// The constraint names are Postgres defaults: <table>_<column>_key.
func mapUserErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		switch pgErr.ConstraintName {
		case "users_email_key":
			return domain.ErrEmailTaken
		case "users_phone_key":
			return domain.ErrPhoneTaken
		case "user_token_expired":
			return domain.ErrInvalidRefreshToken
		}
	}
	return fmt.Errorf("user repo: %w", err)
}
