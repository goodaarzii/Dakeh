package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

type ProductRepo struct {
	q sqlc.Querier
}

func NewProductRepo(q sqlc.Querier) *ProductRepo {
	return &ProductRepo{q: q}
}

func (r *ProductRepo) Create(ctx context.Context, u domain.Product) (domain.Product, error) {
	product, err := r.q.CreateProduct(ctx, sqlc.CreateProductParams{
		CategoryID:  u.CategoryID,
		Name:        u.Name,
		Slug:        u.Slug,
		Description: u.Description,
		Brand:       u.Brand,
		Status:      string(domain.ProductStatus(u.Status)),
	})
	if err != nil {
		return domain.Product{}, err
	}

	return toDomainProduct(product), nil

}

func (r *ProductRepo) Update(ctx context.Context, u domain.Product) (domain.Product, error) {
	updatedProduct, err := r.q.UpdateProduct(ctx, sqlc.UpdateProductParams{
		CategoryID:  u.CategoryID,
		Name:        u.Name,
		Slug:        u.Slug,
		Description: u.Description,
		Brand:       u.Brand,
		Status:      string(domain.ProductStatus(u.Status)),
		ID:          u.ID,
	})

	if err != nil {
		return domain.Product{}, nil
	}

	return toDomainProduct(updatedProduct), nil
}

func (r *ProductRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Product, error) {

	productByID, err := r.q.GetProductByID(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}

	return toDomainProduct(productByID), nil

}

func (r *ProductRepo) Delete(ctx context.Context, uuid2 uuid.UUID) (int64, error) {
	result, err := r.q.SoftDeleteProduct(ctx, uuid2)
	if err != nil {
		return -1, err
	}
	return result, err
}

func (r *ProductRepo) ListProducts(ctx context.Context, offset int, pageSize int) ([]domain.Product, error) {
	products, err := r.q.ListProducts(ctx, sqlc.ListProductsParams{PageOffset: int32(offset), PageSize: int32(pageSize)})
	if err != nil {
		return nil, err
	}
	slog.Info("products in repo 1", int32(pageSize), int32(offset))
	return toDomainProducts(products), nil
}
func (r *ProductRepo) ListProductsByCategory(ctx context.Context, categoryID *uuid.UUID, limit, offset int) ([]domain.Product, error) {
	rows, err := r.q.ListProductsByCategory(ctx, sqlc.ListProductsByCategoryParams{
		CategoryID: *categoryID,
		PageSize:   int32(limit),
		PageOffset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("product repo: list: %w", err)
	}
	return toDomainProducts(rows), nil
}

func toDomainProduct(p sqlc.Product) domain.Product {
	return domain.Product{
		ID:          p.ID,
		CategoryID:  p.CategoryID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Brand:       p.Brand,
		Status:      domain.ProductStatus(p.Status),
		AvgRating:   p.AvgRating,
		RatingCount: int(p.RatingCount),
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toDomainProducts(rows []sqlc.Product) []domain.Product {
	out := make([]domain.Product, len(rows))
	for i, row := range rows {
		out[i] = toDomainProduct(row)
	}
	return out
}
