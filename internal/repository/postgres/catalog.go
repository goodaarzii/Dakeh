package postgres

import (
	"context"
	"encoding/json"
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
	rows, err := r.q.ListActiveCategories(ctx)
	if err != nil {
		return nil, mapCatalogErr(err)
	}
	return toDomainCategories(rows), nil
}

func (r *CatalogRepo) ListActiveVariantsByProduct(ctx context.Context, productID uuid.UUID) ([]domain.Variant, error) {
	rows, err := r.q.ListActiveVariantsByProduct(ctx, productID)
	if err != nil {
		return nil, mapCatalogErr(err)
	}
	return toDomainVariants(rows)
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

func (r *CatalogRepo) GetActiveProductBySlug(ctx context.Context, slug string) (domain.ProductDetail, error) {
	productBySlug, err := r.q.GetActiveProductBySlug(ctx, slug)
	if err != nil {
		return domain.ProductDetail{}, mapCatalogErr(err) // no row -> ErrNotFound -> 404
	}

	variantsByID, err := r.q.ListVariantsByProduct(ctx, productBySlug.ID)
	if err != nil {
		return domain.ProductDetail{}, mapCatalogErr(err)
	}
	var variantIDs = make([]uuid.UUID, 0)
	for _, variant := range variantsByID {
		variantIDs = append(variantIDs, variant.ID)
	}
	stocks, err := r.q.ListInventoryByVariantIDs(ctx, variantIDs)
	if err != nil {
		return domain.ProductDetail{}, mapVariantErr(err)
	}
	tags, err := r.q.ListTagsByProduct(ctx, productBySlug.ID)
	if err != nil {
		return domain.ProductDetail{}, mapVariantErr(err)
	}
	categoryByID, err := r.q.GetCategoryByID(ctx, productBySlug.CategoryID)
	if err != nil {
		return domain.ProductDetail{}, mapCatalogErr(err)
	}

	var listVarWithStocks []domain.VariantWithStock

	for index, _ := range stocks {
		listVarWithStocks = append(listVarWithStocks, domain.VariantWithStock{
			Variant: toDomainVariantV2(variantsByID[index]),
			Stock:   toDomainInventory(stocks[index]),
		})
	}
	return domain.ProductDetail{
		Product:  toDomainProduct(productBySlug),
		Category: toDomainCategory(categoryByID),
		Variants: listVarWithStocks,
		Tags:     toDomainTags(tags),
	}, nil

}

func toDomainVariantV2(v sqlc.ProductVariant) domain.Variant {
	attrs := map[string]string{}
	if len(v.Attributes) > 0 {
		if err := json.Unmarshal(v.Attributes, &attrs); err != nil {
			// Fails if someone stored a non-string value (like {"size": 42}) by hand.
			return domain.Variant{}
		}
	}

	return domain.Variant{
		ID:             v.ID,
		ProductID:      v.ProductID,
		SKU:            v.Sku,
		Name:           v.Name,
		Price:          v.Price,
		CompareAtPrice: v.CompareAtPrice,
		Attributes:     attrs,
		WeightGrams:    int(v.WeightGrams),
		IsActive:       v.IsActive,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}
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
