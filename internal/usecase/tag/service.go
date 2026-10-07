package tag

import (
	"context"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, address domain.Tag) (domain.Tag, error)
	Update(ctx context.Context, address domain.Tag) (domain.Tag, error)
	Delete(ctx context.Context, id uuid.UUID) (int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Tag, error)
	AddTagToProduct(ctx context.Context, id uuid.UUID, tagID []uuid.UUID) error
	RemoveTagFromProduct(ctx context.Context, id uuid.UUID, tagID uuid.UUID) (int64, error)
	ClearProductTags(ctx context.Context, id uuid.UUID) error
	ListTags(ctx context.Context) ([]domain.Tag, error)
	ListTagsByProduct(ctx context.Context, id uuid.UUID) ([]domain.Tag, error)
}

type TokenIssuer interface {
	GetToken(userID uuid.UUID) (string, error)
}

type Service struct {
	repo        Repository
	tokenIssuer TokenIssuer
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, address domain.Tag) (domain.Tag, error) {
	return s.repo.Create(ctx, address)
}

func (s *Service) Update(ctx context.Context, address domain.Tag) (domain.Tag, error) {
	return s.repo.Update(ctx, address)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) (int64, error) {
	return s.repo.Delete(ctx, id)
}
func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (domain.Tag, error) {
	return s.repo.GetByID(ctx, id)
}
func (s *Service) AddTagToProduct(ctx context.Context, id uuid.UUID, tagID []uuid.UUID) error {
	return s.repo.AddTagToProduct(ctx, id, tagID)
}

func (s *Service) ClearProductTags(ctx context.Context, id uuid.UUID) error {
	return s.repo.ClearProductTags(ctx, id)
}

func (s *Service) ListTags(ctx context.Context) ([]domain.Tag, error) {
	return s.repo.ListTags(ctx)
}

func (s *Service) ListTagsByProduct(ctx context.Context, productID uuid.UUID) ([]domain.Tag, error) {
	return s.repo.ListTagsByProduct(ctx, productID)
}

func (s *Service) RemoveTagFromProduct(ctx context.Context, id uuid.UUID, tagID uuid.UUID) (int64, error) {
	return s.repo.RemoveTagFromProduct(ctx, id, tagID)
}
