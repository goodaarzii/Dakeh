package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// withTx runs fn inside one database transaction.
// If fn returns an error, everything is rolled back. If it returns nil, everything is committed.
// You will use this again for checkout (order + items + stock in one transaction).
func withTx(ctx context.Context, pool *pgxpool.Pool, fn func(q *sqlc.Queries) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	// After a successful Commit this does nothing. After an error it rolls back.
	defer tx.Rollback(ctx)

	if err := fn(sqlc.New(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

const (
	pgForeignKeyViolation = "23503"
	pgCheckViolation      = "23514"
)

// pgConstraint returns the Postgres error code and the constraint name ("" and "" if it is not a Postgres error).
func pgConstraint(err error) (code, constraint string) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code, pgErr.ConstraintName
	}
	return "", ""
}
