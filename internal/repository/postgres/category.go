package postgres

import (
	"context"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

type CategoryRepo struct {
	q sqlc.Querier
}

func NewCategoryRepo(q sqlc.Querier) *CategoryRepo {
	return &CategoryRepo{q: q}
}

func (r *CategoryRepo) Create(ctx context.Context, u domain.Category) (domain.Category, error) {
	category, err := r.q.CreateCategory(ctx, sqlc.CreateCategoryParams{
		ParentID:    u.ParentID,
		Name:        u.Name,
		Slug:        u.Slug,
		Description: u.Description,
		ImageUrl:    u.ImageURL,
		SortOrder:   int32(u.SortOrder),
		IsActive:    u.IsActive,
	})
	if err != nil {
		return domain.Category{}, err
	}

	return toDomainCategory(category), nil

}

func (r *CategoryRepo) CategoryList(ctx context.Context) ([]domain.Category, error) {
	categories, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainCategories(categories), nil

}

func (r *CategoryRepo) FindByID(ctx context.Context, id uuid.UUID) (domain.Category, error) {
	categoryByID, err := r.q.GetCategoryByID(ctx, id)
	if err != nil {
		return domain.Category{}, err
	}

	return toDomainCategory(categoryByID), nil

}

func (r *CategoryRepo) Update(ctx context.Context, category domain.Category) (domain.Category, error) {
	updateCategory, err := r.q.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:          category.ID,
		ParentID:    category.ParentID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		ImageUrl:    category.ImageURL,
		SortOrder:   int32(category.SortOrder),
		IsActive:    category.IsActive,
	})
	if err != nil {
		return domain.Category{}, err
	}
	return toDomainCategory(updateCategory), nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	deleteCategory, err := r.q.DeleteCategory(ctx, id)
	if err != nil {
		return -1, err
	}
	return deleteCategory, nil
}

func toDomainCategory(p sqlc.Category) domain.Category {
	return domain.Category{
		ID:          p.ID,
		ParentID:    p.ParentID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		ImageURL:    p.ImageUrl,
		SortOrder:   int(p.SortOrder),
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toDomainCategories(rows []sqlc.Category) []domain.Category {
	out := make([]domain.Category, len(rows))
	for i, row := range rows {
		out[i] = toDomainCategory(row)
	}
	return out
}
