package favorite

import (
	"context"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

// Repository is the persistence contract required by the favorites use case.
type Repository interface {
	AddFavorite(ctx context.Context, favorite domain.Favorite) (domain.Favorite, error)
	RemoveFavorite(ctx context.Context, userID, productID uuid.UUID) error
	ListFavoriteProducts(ctx context.Context, userID uuid.UUID) ([]domain.FavoriteProduct, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Add is safe to repeat: the repository returns the already-saved favorite if it exists.
func (s *Service) Add(ctx context.Context, userID, productID uuid.UUID) (domain.Favorite, error) {
	favorite, err := domain.NewFavorite(userID, productID)
	if err != nil {
		return domain.Favorite{}, err
	}
	return s.repo.AddFavorite(ctx, favorite)
}

// Remove is idempotent; removing a product that is not currently favorited succeeds.
func (s *Service) Remove(ctx context.Context, userID, productID uuid.UUID) error {
	favorite, err := domain.NewFavorite(userID, productID)
	if err != nil {
		return err
	}
	return s.repo.RemoveFavorite(ctx, favorite.UserID, favorite.ProductID)
}

func (s *Service) List(ctx context.Context, userID uuid.UUID) ([]domain.FavoriteProduct, error) {
	if userID == uuid.Nil {
		return nil, domain.ErrInvalidInput
	}
	return s.repo.ListFavoriteProducts(ctx, userID)
}
