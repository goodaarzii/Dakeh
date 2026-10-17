package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
)

// ImageRepo needs the pool because Create, SetPrimary and Delete use transactions.
// In routes.go: postgres.NewImageRepo(a.DB)
type ImageRepo struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewImageRepo(pool *pgxpool.Pool) *ImageRepo {
	return &ImageRepo{pool: pool, q: sqlc.New(pool)}
}

// Create saves the image row. The FIRST image of a product becomes primary automatically.
// Everything runs in one transaction: insert, check the variant, promote, re-read.
func (r *ImageRepo) Create(ctx context.Context, img domain.ProductImage) (domain.ProductImage, error) {
	var result domain.ProductImage

	err := withTx(ctx, r.pool, func(q *sqlc.Queries) error {
		// A variant photo must belong to a variant OF THIS product.
		if img.VariantID != nil {
			v, err := q.GetVariantByID(ctx, *img.VariantID)
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.ErrInvalidInput
			}
			if err != nil {
				return err
			}
			if v.ProductID != img.ProductID {
				return domain.ErrInvalidInput
			}
		}

		row, err := q.CreateProductImage(ctx, sqlc.CreateProductImageParams{
			ProductID:  img.ProductID,
			VariantID:  img.VariantID,
			StorageKey: img.StorageKey,
			Url:        img.URL, // if your generated field is called Url, rename it here
			AltText:    img.AltText,
			SortOrder:  int32(img.SortOrder),
			IsPrimary:  false, // PromoteNextImage below decides
		})
		if err != nil {
			return err
		}

		// Makes the first remaining image primary, but only if the product has no primary yet.
		if err := q.PromoteNextImage(ctx, img.ProductID); err != nil {
			return err
		}

		// Read it again: is_primary may have just changed.
		row, err = q.GetProductImageByID(ctx, row.ID)
		if err != nil {
			return err
		}
		result = toDomainImage(row)
		return nil
	})
	if err != nil {
		return domain.ProductImage{}, mapImageErr(err)
	}
	return result, nil
}

func (r *ImageRepo) CountByProduct(ctx context.Context, productID uuid.UUID) (int, error) {
	n, err := r.q.CountImagesByProduct(ctx, productID)
	if err != nil {
		return 0, mapImageErr(err)
	}
	return int(n), nil
}

func (r *ImageRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.ProductImage, error) {
	row, err := r.q.GetProductImageByID(ctx, id)
	if err != nil {
		return domain.ProductImage{}, mapImageErr(err)
	}
	return toDomainImage(row), nil
}

func (r *ImageRepo) ListByProduct(ctx context.Context, productID uuid.UUID) ([]domain.ProductImage, error) {
	rows, err := r.q.ListImagesByProduct(ctx, productID)
	if err != nil {
		return nil, mapImageErr(err)
	}
	return toDomainImages(rows), nil
}

func (r *ImageRepo) Update(ctx context.Context, id uuid.UUID, altText string, sortOrder int) (domain.ProductImage, error) {
	row, err := r.q.UpdateProductImage(ctx, sqlc.UpdateProductImageParams{
		ID:        id,
		AltText:   altText,
		SortOrder: int32(sortOrder),
	})
	if err != nil {
		return domain.ProductImage{}, mapImageErr(err) // never return (empty, nil)
	}
	return toDomainImage(row), nil
}

// SetPrimary makes imageID the primary image of the product.
// Two steps in ONE transaction: clear the old primary (the unique index allows only one), set the new one.
// If the image does not belong to this product, the transaction rolls back and the old primary stays.
func (r *ImageRepo) SetPrimary(ctx context.Context, productID, imageID uuid.UUID) error {
	err := withTx(ctx, r.pool, func(q *sqlc.Queries) error {
		if err := q.ClearPrimaryImage(ctx, productID); err != nil {
			return err
		}
		n, err := q.SetPrimaryImage(ctx, sqlc.SetPrimaryImageParams{ID: imageID, ProductID: productID})
		if err != nil {
			return err
		}
		if n == 0 {
			return domain.ErrNotFound
		}
		return nil
	})
	return mapImageErr(err)
}

// Delete removes the row and returns it (the usecase needs StorageKey to delete the file).
// If it was the primary image, the next image is promoted, in the same transaction.
func (r *ImageRepo) Delete(ctx context.Context, id uuid.UUID) (domain.ProductImage, error) {
	var deleted domain.ProductImage

	err := withTx(ctx, r.pool, func(q *sqlc.Queries) error {
		row, err := q.DeleteProductImage(ctx, id)
		if err != nil {
			return err
		}
		if row.IsPrimary {
			if err := q.PromoteNextImage(ctx, row.ProductID); err != nil {
				return err
			}
		}
		deleted = toDomainImage(row)
		return nil
	})
	if err != nil {
		return domain.ProductImage{}, mapImageErr(err)
	}
	return deleted, nil
}

func toDomainImage(i sqlc.ProductImage) domain.ProductImage {
	return domain.ProductImage{
		ID:         i.ID,
		ProductID:  i.ProductID,
		VariantID:  i.VariantID,
		StorageKey: i.StorageKey,
		URL:        i.Url, // if your generated field is called Url, rename it here
		AltText:    i.AltText,
		SortOrder:  int(i.SortOrder),
		IsPrimary:  i.IsPrimary,
		CreatedAt:  i.CreatedAt,
	}
}

// toDomainImages is also used by the catalog repo (product detail page).
func toDomainImages(rows []sqlc.ProductImage) []domain.ProductImage {
	out := make([]domain.ProductImage, len(rows))
	for i, row := range rows {
		out[i] = toDomainImage(row)
	}
	return out
}

func mapImageErr(err error) error {
	if err == nil {
		return nil
	}
	// Already domain errors (returned from inside a transaction): pass them through.
	if errors.Is(err, domain.ErrNotFound) || errors.Is(err, domain.ErrInvalidInput) {
		return err
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" { // foreign key violation
		switch pgErr.ConstraintName {
		case "product_images_product_id_fkey":
			return domain.ErrNotFound // the product does not exist
		case "product_images_variant_id_fkey":
			return domain.ErrInvalidInput
		}
	}
	return fmt.Errorf("image repo: %w", err)
}
