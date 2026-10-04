package postgres

import (
	"context"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
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
