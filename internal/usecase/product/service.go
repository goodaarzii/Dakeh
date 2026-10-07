package product

import (
	"context"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, address domain.Product) (domain.Product, error)
	Update(ctx context.Context, address domain.Product) (domain.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Product, error)
	Delete(ctx context.Context, id uuid.UUID) (int64, error)
	ListProducts(ctx context.Context, offset int, pageSize int) ([]domain.Product, error)
	ListProductsByCategory(ctx context.Context, categoryID *uuid.UUID, offset int, pageSize int) ([]domain.Product, error)
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

func (s *Service) Create(ctx context.Context, u domain.Product) (domain.Product, error) {
	product, err := s.repo.Create(ctx, u)
	if err != nil {
		return domain.Product{}, err
	}

	return product, nil

}
func (s *Service) Update(ctx context.Context, u domain.Product) (domain.Product, error) {
	product, err := s.repo.Update(ctx, u)
	if err != nil {
		return domain.Product{}, err
	}

	return product, nil

}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (domain.Product, error) {
	product, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Product{}, err
	}
	return product, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	result, err := s.repo.Delete(ctx, id)
	if err != nil {
		return -1, err
	}
	return result, nil
}

func (s *Service) ListProducts(ctx context.Context, offset int, pageSize int) ([]domain.Product, error) {
	products, err := s.repo.ListProducts(ctx, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func (s *Service) ListProductsByCategory(ctx context.Context, categoryID *uuid.UUID, offset int, pageSize int) ([]domain.Product, error) {
	products, err := s.repo.ListProductsByCategory(ctx, categoryID, offset, pageSize)
	if err != nil {
		return nil, err
	}
	return products, nil
}
