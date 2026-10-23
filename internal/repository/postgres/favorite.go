package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
)

type FavoriteRepo struct {
	q sqlc.Querier
}

func NewFavoriteRepo(q sqlc.Querier) *FavoriteRepo {
	return &FavoriteRepo{q: q}
}

func (r *FavoriteRepo) AddFavorite(ctx context.Context, favorite domain.Favorite) (domain.Favorite, error) {
	rows, err := r.q.AddFavorite(ctx, sqlc.AddFavoriteParams{
		UserID:    favorite.UserID,
		ProductID: favorite.ProductID,
	})
	if err != nil {
		return domain.Favorite{}, fmt.Errorf("favorite repo: add: %w", err)
	}

	// If the INSERT affected no rows, it was either a repeat PUT or the product was not
	// publicly available. Check which case it is so only repeats succeed idempotently.
	if rows == 0 {
		exists, err := r.q.IsFavorite(ctx, sqlc.IsFavoriteParams{
			UserID:    favorite.UserID,
			ProductID: favorite.ProductID,
		})
		if err != nil {
			return domain.Favorite{}, fmt.Errorf("favorite repo: check existing: %w", err)
		}
		if !exists {
			return domain.Favorite{}, domain.ErrNotFound
		}
	}

	createdAt, err := r.q.GetFavoriteCreatedAt(ctx, sqlc.GetFavoriteCreatedAtParams{
		UserID:    favorite.UserID,
		ProductID: favorite.ProductID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Favorite{}, domain.ErrNotFound
		}
		return domain.Favorite{}, fmt.Errorf("favorite repo: get saved favorite: %w", err)
	}
	favorite.CreatedAt = createdAt
	return favorite, nil
}

func (r *FavoriteRepo) RemoveFavorite(ctx context.Context, userID, productID uuid.UUID) error {
	_, err := r.q.RemoveFavorite(ctx, sqlc.RemoveFavoriteParams{UserID: userID, ProductID: productID})
	if err != nil {
		return fmt.Errorf("favorite repo: remove: %w", err)
	}
	return nil
}

func (r *FavoriteRepo) ListFavoriteProducts(ctx context.Context, userID uuid.UUID) ([]domain.FavoriteProduct, error) {
	rows, err := r.q.ListFavoriteProducts(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("favorite repo: list: %w", err)
	}

	favorites := make([]domain.FavoriteProduct, len(rows))
	for i, row := range rows {
		favorites[i] = domain.FavoriteProduct{
			Product: domain.ProductCard{
				ID:          row.ProductID,
				CategoryID:  row.CategoryID,
				Name:        row.Name,
				Slug:        row.Slug,
				Brand:       row.Brand,
				MinPrice:    row.MinPrice,
				InStock:     row.InStock,
				ImageURL:    row.ImageUrl,
				AvgRating:   row.AvgRating,
				RatingCount: int(row.RatingCount),
			},
			CreatedAt: row.FavoritedAt,
		}
	}
	return favorites, nil
}
