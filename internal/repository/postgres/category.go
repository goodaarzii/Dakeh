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
		return domain.Category{}, mapCategoryErr(err)
	}

	return toDomainCategory(category), nil

}

func (r *CategoryRepo) CategoryList(ctx context.Context) ([]domain.Category, error) {
	categories, err := r.q.ListCategories(ctx)
	if err != nil {
		return nil, mapCategoryErr(err)
	}
	return toDomainCategories(categories), nil

}

func (r *CategoryRepo) FindByID(ctx context.Context, id uuid.UUID) (domain.Category, error) {
	categoryByID, err := r.q.GetCategoryByID(ctx, id)
	if err != nil {
		return domain.Category{}, mapCategoryErr(err)
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
		return domain.Category{}, mapCategoryErr(err)
	}
	return toDomainCategory(updateCategory), nil
}

func (r *CategoryRepo) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	deleteCategory, err := r.q.DeleteCategory(ctx, id)
	if err != nil {
		if code, _ := pgConstraint(err); code == pgForeignKeyViolation {
			return 0, domain.ErrCategoryInUse // it still has products or child categories
		}
		return 0, mapCategoryErr(err)
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

func mapCategoryErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	switch code, constraint := pgConstraint(err); code {
	case pgUniqueViolation:
		if constraint == "categories_slug_key" {
			return domain.ErrSlugTaken
		}
	case pgForeignKeyViolation:
		if constraint == "categories_parent_id_fkey" {
			return domain.ErrInvalidInput // the parent category does not exist
		}
	case pgCheckViolation:
		return domain.ErrInvalidInput // for example a category that is its own parent
	}
	return fmt.Errorf("category repo: %w", err)
}
