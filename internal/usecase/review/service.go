package review

import (
	"context"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// Repository is the persistence contract needed by the review use case.
type Repository interface {
	ListByProductID(ctx context.Context, productID uuid.UUID, limit, offset int) ([]domain.Review, int, error)
	GetMineForProduct(ctx context.Context, userID, productID uuid.UUID) (domain.Review, error)
	Create(ctx context.Context, review domain.Review) (domain.Review, error)
	UpdateMine(ctx context.Context, userID, reviewID uuid.UUID, draft domain.ReviewDraft) (domain.Review, error)
	DeleteMine(ctx context.Context, userID, reviewID uuid.UUID) error
	ListOfReviews(ctx context.Context, userid uuid.UUID, limit, offset int) ([]domain.Review, int, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ListByProductID(ctx context.Context, productID uuid.UUID, page, pageSize int) ([]domain.Review, int, bool, error) {
	if productID == uuid.Nil {
		return nil, 0, false, domain.ErrInvalidInput
	}
	if page < 1 {
		page = 1
	}
	// Keep the SQL offset within int32 because that is PostgreSQL LIMIT/OFFSET's generated type.
	if page > 1_000_000 {
		return nil, 0, false, domain.ErrInvalidInput
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	items, total, err := s.repo.ListByProductID(ctx, productID, pageSize+1, (page-1)*pageSize)
	if err != nil {
		return nil, 0, false, err
	}
	hasMore := len(items) > pageSize
	if hasMore {
		items = items[:pageSize]
	}
	return items, total, hasMore, nil
}

func (s *Service) GetMineForProduct(ctx context.Context, userID, productID uuid.UUID) (domain.Review, error) {
	if userID == uuid.Nil || productID == uuid.Nil {
		return domain.Review{}, domain.ErrInvalidInput
	}
	return s.repo.GetMineForProduct(ctx, userID, productID)
}

func (s *Service) Create(ctx context.Context, userID, productID uuid.UUID, rating int, title, body string) (domain.Review, error) {
	review, err := domain.NewReview(userID, productID, rating, title, body)
	if err != nil {
		return domain.Review{}, err
	}
	return s.repo.Create(ctx, review)
}

func (s *Service) UpdateMine(ctx context.Context, userID, reviewID uuid.UUID, rating int, title, body string) (domain.Review, error) {
	if userID == uuid.Nil || reviewID == uuid.Nil {
		return domain.Review{}, domain.ErrInvalidInput
	}
	draft, err := domain.NewReviewDraft(rating, title, body)
	if err != nil {
		return domain.Review{}, err
	}
	return s.repo.UpdateMine(ctx, userID, reviewID, draft)
}

func (s *Service) DeleteMine(ctx context.Context, userID, reviewID uuid.UUID) error {
	if userID == uuid.Nil || reviewID == uuid.Nil {
		return domain.ErrInvalidInput
	}
	return s.repo.DeleteMine(ctx, userID, reviewID)
}
