package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
)

const pgReviewUniqueViolation = "23505"

type ReviewRepo struct {
	q sqlc.Querier
}

func NewReviewRepo(q sqlc.Querier) *ReviewRepo {
	return &ReviewRepo{q: q}
}
func (r *ReviewRepo) ListOfReviews(ctx context.Context, userid uuid.UUID, limit, offset int) ([]domain.Review, int, error) {
	rows, err := r.q.ListReviews(ctx, sqlc.ListReviewsParams{
		UserID: userid,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, 0, mapReviewErr(err)
	}

	items := make([]domain.Review, len(rows))
	for i, row := range rows {
		items[i] = domain.Review{
			ID:           row.ID,
			ProductID:    row.ProductID,
			UserID:       row.UserID,
			ReviewerName: row.ReviewerName,
			Rating:       int(row.Rating),
			Title:        row.Title,
			Body:         row.Body,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		}
	}
	return items, len(items), nil
}

func (r *ReviewRepo) ListByProductID(ctx context.Context, productID uuid.UUID, limit, offset int) ([]domain.Review, int, error) {
	publicID, err := r.q.GetPublicProductIDByID(ctx, productID)
	if err != nil {
		return nil, 0, mapReviewErr(err)
	}
	rows, err := r.q.ListReviewsByProduct(ctx, sqlc.ListReviewsByProductParams{
		ProductID: publicID, PageSize: int32(limit), PageOffset: int32(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("review repo: list: %w", err)
	}
	count, err := r.q.CountReviewsByProduct(ctx, publicID)
	if err != nil {
		return nil, 0, fmt.Errorf("review repo: count: %w", err)
	}
	items := make([]domain.Review, len(rows))
	for i, row := range rows {
		items[i] = domain.Review{
			ID: row.ID, ProductID: row.ProductID, UserID: row.UserID,
			ReviewerName: row.ReviewerName, Rating: int(row.Rating),
			Title: row.Title, Body: row.Body,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
	}
	return items, int(count), nil
}

func (r *ReviewRepo) GetMineForProduct(ctx context.Context, userID, productID uuid.UUID) (domain.Review, error) {
	row, err := r.q.GetMyReviewForProduct(ctx, sqlc.GetMyReviewForProductParams{
		ProductID: productID, UserID: userID,
	})
	if err != nil {
		return domain.Review{}, mapReviewErr(err)
	}
	return domain.Review{
		ID: row.ID, ProductID: row.ProductID, UserID: row.UserID,
		ReviewerName: row.ReviewerName, Rating: int(row.Rating),
		Title: row.Title, Body: row.Body,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *ReviewRepo) Create(ctx context.Context, review domain.Review) (domain.Review, error) {
	row, err := r.q.CreateProductReview(ctx, sqlc.CreateProductReviewParams{
		ProductID: review.ProductID, UserID: review.UserID,
		Rating: int16(review.Rating), Title: review.Title, Body: review.Body,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Review{}, domain.ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgReviewUniqueViolation && pgErr.ConstraintName == "uq_reviews_product_user" {
			return domain.Review{}, domain.ErrConflict
		}
		return domain.Review{}, fmt.Errorf("review repo: create: %w", err)
	}
	return domain.Review{
		ID: row.ID, ProductID: row.ProductID, UserID: row.UserID,
		ReviewerName: row.ReviewerName, Rating: int(row.Rating),
		Title: row.Title, Body: row.Body,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *ReviewRepo) UpdateMine(ctx context.Context, userID, reviewID uuid.UUID, draft domain.ReviewDraft) (domain.Review, error) {
	row, err := r.q.UpdateMyReview(ctx, sqlc.UpdateMyReviewParams{
		ReviewID: reviewID, UserID: userID,
		Rating: int16(draft.Rating), Title: draft.Title, Body: draft.Body,
	})
	if err != nil {
		return domain.Review{}, mapReviewErr(err)
	}
	return domain.Review{
		ID: row.ID, ProductID: row.ProductID, UserID: row.UserID,
		ReviewerName: row.ReviewerName, Rating: int(row.Rating),
		Title: row.Title, Body: row.Body,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *ReviewRepo) DeleteMine(ctx context.Context, userID, reviewID uuid.UUID) error {
	rows, err := r.q.DeleteMyReview(ctx, sqlc.DeleteMyReviewParams{ReviewID: reviewID, UserID: userID})
	if err != nil {
		return fmt.Errorf("review repo: delete: %w", err)
	}
	if rows == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func mapReviewErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	return fmt.Errorf("review repo: %w", err)
}
