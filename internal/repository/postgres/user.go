package postgres

import (
	"context"
	"errors"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/jackc/pgx/v5"
)

type UserRepo struct{ q sqlc.Querier }

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrNotFound
	}

}
