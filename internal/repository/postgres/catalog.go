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

type CatalogRepo struct {
	q sqlc.Querier
}

func NewCatalogRepo(q sqlc.Querier) *CatalogRepo {
	return &CatalogRepo{q: q}
}

func (r *CatalogRepo) ListActiveCategories(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.q.ListActiveCategories(ctx)
	if err != nil {
		return nil, mapCatalogErr(err)
	}
	return toDomainCategories(rows), nil
}

// ListActiveVariantsWithStock returns the sellable variants of a product together with their stock.
// One query: variant -> inventory is 1:1, so the JOIN cannot repeat a variant.
func (r *CatalogRepo) ListActiveVariantsWithStock(ctx context.Context, productID uuid.UUID) ([]domain.VariantWithStock, error) {
	rows, err := r.q.ListActiveVariantsWithStock(ctx, productID)
	if err != nil {
		return nil, mapCatalogErr(err)
	}

	out := make([]domain.VariantWithStock, len(rows))
	for i, row := range rows {
		v, err := toDomainVariant(row.ProductVariant)
		if err != nil {
			return nil, mapCatalogErr(err)
		}
		out[i] = domain.VariantWithStock{
			Variant: v,
			Stock: domain.Inventory{
				VariantID:         v.ID,
				QuantityOnHand:    int(row.QuantityOnHand),
				QuantityReserved:  int(row.QuantityReserved),
				LowStockThreshold: int(row.LowStockThreshold),
			},
		}
	}
	return out, nil
}

func (r *CatalogRepo) ListPublicProducts(ctx context.Context, q domain.ProductQuery) ([]domain.ProductCard, error) {
	categoryIDs := q.CategoryIDs
	if categoryIDs == nil {
		categoryIDs = []uuid.UUID{}
	}

	rows, err := r.q.ListPublicProducts(ctx, sqlc.ListPublicProductsParams{
		CategoryIds: categoryIDs,
		TagSlug:     q.TagSlug,
		Search:      q.Search,
		Sort:        string(q.Sort),
		PageSize:    int32(q.Limit),
		PageOffset:  int32(q.Offset),
	})
	if err != nil {
		return nil, mapCatalogErr(err)
	}
	return toDomainPublicProducts(rows), nil
}

func (r *CatalogRepo) GetTags(ctx context.Context) ([]domain.Tag, error) {
	tags, err := r.q.ListTags(ctx)
	if err != nil {
		return nil, mapCatalogErr(err)
	}
	return toDomainTags(tags), nil
}

// GetActiveProductBySlug builds the whole product page: 4 queries, always 4,
// however many variants the product has.
func (r *CatalogRepo) GetActiveProductBySlug(ctx context.Context, slug string) (domain.ProductDetail, error) {
	// 1. the product (no row -> ErrNotFound -> 404)
	product, err := r.q.GetActiveProductBySlug(ctx, slug)
	if err != nil {
		return domain.ProductDetail{}, mapCatalogErr(err)
	}

	// 2. sellable variants and their stock, already matched by the database
	variants, err := r.ListActiveVariantsWithStock(ctx, product.ID)
	if err != nil {
		return domain.ProductDetail{}, err
	}
	// The product list hides products without sellable variants, so the detail page does too.
	if len(variants) == 0 {
		return domain.ProductDetail{}, domain.ErrNotFound
	}

	// 3. tags
	tagRows, err := r.q.ListTagsByProduct(ctx, product.ID)
	if err != nil {
		return domain.ProductDetail{}, mapCatalogErr(err)
	}

	// 4. category
	categoryRow, err := r.q.GetCategoryByID(ctx, product.CategoryID)
	if err != nil {
		return domain.ProductDetail{}, mapCatalogErr(err)
	}

	return domain.ProductDetail{
		Product:  toDomainProduct(product),
		Category: toDomainCategory(categoryRow),
		Variants: variants,
		Tags:     toDomainTags(tagRows),
	}, nil
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
