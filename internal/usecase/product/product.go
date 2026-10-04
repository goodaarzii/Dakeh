package product

import (
	"context"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, address domain.Product) (domain.Product, error)
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
