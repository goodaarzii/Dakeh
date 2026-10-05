package category

import (
	"context"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, address domain.Category) (domain.Category, error)
	Update(ctx context.Context, address domain.Category) (domain.Category, error)
	CategoryList(ctx context.Context) ([]domain.Category, error)
	FindByID(ctx context.Context, id uuid.UUID) (domain.Category, error)
	Delete(ctx context.Context, id uuid.UUID) (int64, error)
}

type TokenIssuer interface {
	GetToken(userID uuid.UUID) (string, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, address domain.Category) (domain.Category, error) {
	return s.repo.Create(ctx, address)
}

func (s *Service) CategoryList(ctx context.Context) ([]domain.Category, error) {
	return s.repo.CategoryList(ctx)
}

func (s *Service) FindByID(ctx context.Context, id uuid.UUID) (domain.Category, error) {
	return s.repo.FindByID(ctx, id)
}
func (s *Service) Update(ctx context.Context, address domain.Category) (domain.Category, error) {
	return s.repo.Update(ctx, address)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	return s.repo.Delete(ctx, id)
}
