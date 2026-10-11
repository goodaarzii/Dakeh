package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CatalogRepo struct {
	q sqlc.Querier
}

func NewCatalogRepo(q sqlc.Querier) *CatalogRepo {
	return &CatalogRepo{q: q}
}

func (r *CatalogRepo) ListActiveCategories(ctx context.Context) ([]domain.Category, error) {
	categories, err := r.q.ListActiveCategories(ctx)
	if err != nil {
		return nil, mapCatalogErr(err)
	}
	return toDomainCategories(categories), nil

}
func (r *CatalogRepo) GetActiveProductBySlug(ctx context.Context, slug string) (domain.Product, error) {
	bySlug, err := r.q.GetActiveProductBySlug(ctx, slug)
	if err != nil {
		return domain.Product{}, mapCatalogErr(err)
	}
	return toDomainProduct(bySlug), nil

}

func (r *CatalogRepo) ListActiveVariantsByProduct(ctx context.Context, id uuid.UUID) ([]domain.Variant, error) {
	product, err := r.q.ListActiveVariantsByProduct(ctx, id)
	if err != nil {
		return []domain.Variant{}, mapCatalogErr(err)
	}

	return toDomainVariants(product)

}

func (r *CatalogRepo) ListPublicProducts(ctx context.Context, sort, tagSlug, search string, offSet, pageSize int, catIds []uuid.UUID) ([]domain.ProductCard, error) {

	categoryIDs := q.CategoryIDs
	if categoryIDs == nil {
		categoryIDs = []uuid.UUID{} // an empty array, not NULL
	}
	products, err := r.q.ListPublicProducts(ctx, sqlc.ListPublicProductsParams{
		Sort:        sort,
		PageOffset:  int32(offSet),
		PageSize:    int32(pageSize),
		CategoryIds: catIds,
		TagSlug:     tagSlug,
		Search:      search,
	})
	if err != nil {
		return []domain.ProductCard{}, mapCatalogErr(err)
	}
	return toDomainPublicProducts(products), nil
}

func toDomainPublicProducts(rows []sqlc.ListPublicProductsRow) []domain.ProductCard {
	out := make([]domain.ProductCard, len(rows))

	for i, row := range rows {
		out[i] = toDomainPublicProduct(row)
	}
	return out
}

func toDomainPublicProduct(row sqlc.ListPublicProductsRow) domain.ProductCard {
	return domain.ProductCard{
		ID:          row.ID,
		CategoryID:  row.CategoryID,
		Name:        row.Name,
		Slug:        row.Slug,
		Brand:       row.Brand,
		MinPrice:    row.MinPrice,
		InStock:     row.InStock,
		AvgRating:   row.AvgRating,
		RatingCount: int(row.RatingCount),
	}
}

func mapCatalogErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return fmt.Errorf("catalog repo: %w", err)
}
