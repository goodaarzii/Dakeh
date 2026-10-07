package postgres

import (
	"context"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

type TagRepo struct {
	q sqlc.Querier
}

func NewTagRepo(q sqlc.Querier) *TagRepo {
	return &TagRepo{q: q}
}

func (t *TagRepo) Create(ctx context.Context, tag domain.Tag) (domain.Tag, error) {
	createTag, err := t.q.CreateTag(ctx, sqlc.CreateTagParams{
		Name: tag.Name,
		Slug: tag.Slug,
	})
	if err != nil {
		return domain.Tag{}, err
	}
	return toDomainTag(createTag), nil
}

func (t *TagRepo) Update(ctx context.Context, tag domain.Tag) (domain.Tag, error) {
	updated, err := t.q.UpdateTag(ctx, sqlc.UpdateTagParams{
		Name: tag.Name,
		Slug: tag.Slug,
	})
	if err != nil {
		return domain.Tag{}, nil
	}
	return toDomainTag(updated), nil
}

func (t *TagRepo) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	tag, err := t.q.DeleteTag(ctx, id)
	if err != nil {
		return -1, err
	}
	return tag, err
}

func (t *TagRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Tag, error) {
	tagByID, err := t.q.GetTagByID(ctx, id)
	if err != nil {
		return domain.Tag{}, err
	}
	return toDomainTag(tagByID), nil
}

func (t *TagRepo) AddTagToProduct(ctx context.Context, productID uuid.UUID, tagIDs []uuid.UUID) error {

	if err := t.ClearProductTags(ctx, productID); err != nil {
		return err
	}
	if len(tagIDs) > 0 {
		if err := t.q.AddTagsToProduct(ctx, sqlc.AddTagsToProductParams{ProductID: productID, TagIds: tagIDs}); err != nil {
			return err
		}
	}
	return nil
}

func (t *TagRepo) ClearProductTags(ctx context.Context, productsID uuid.UUID) error {
	err := t.q.ClearProductTags(ctx, productsID)
	if err != nil {
		return err
	}
	return nil
}

func (t *TagRepo) ListTags(ctx context.Context) ([]domain.Tag, error) {
	tags, err := t.q.ListTags(ctx)
	if err != nil {
		return nil, err
	}
	return toDomainTags(tags), nil
}

func (t *TagRepo) ListTagsByProduct(ctx context.Context, productsIDs uuid.UUID) ([]domain.Tag, error) {
	listTagsByProduct, err := t.q.ListTagsByProduct(ctx, productsIDs)
	if err != nil {
		return nil, err
	}
	return toDomainTags(listTagsByProduct), nil
}

func (t *TagRepo) RemoveTagFromProduct(ctx context.Context, productsID uuid.UUID, tagID uuid.UUID) (int64, error) {
	isRemoved, err := t.q.RemoveTagFromProduct(ctx, sqlc.RemoveTagFromProductParams{
		ProductID: productsID,
		TagID:     tagID,
	})
	if err != nil {
		return -1, err
	}
	return isRemoved, nil
}

func toDomainTag(tag sqlc.Tag) domain.Tag {
	return domain.Tag{
		ID:        tag.ID,
		Name:      tag.Name,
		Slug:      tag.Slug,
		CreatedAt: tag.CreatedAt,
		UpdatedAt: tag.UpdatedAt,
	}
}

func toDomainTags(rows []sqlc.Tag) []domain.Tag {
	out := make([]domain.Tag, len(rows))
	for i, row := range rows {
		out[i] = toDomainTag(row)
	}
	return out
}
