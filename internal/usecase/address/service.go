package address

import (
	"context"
	"time"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, address domain.Address) (domain.Address, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (domain.Address, error)
	GetAllAddresses(ctx context.Context, userID uuid.UUID) ([]domain.Address, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (int64, error)
	update(ctx context.Context, u domain.Address) (domain.Address, error)
}

type TokenIssuer interface {
	GetToken(userID uuid.UUID) (string, error)
}

type Service struct {
	repo   Repository
	tokens TokenIssuer
}

func NewService(repo Repository, tokens TokenIssuer) *Service {
	return &Service{repo: repo, tokens: tokens}
}

type addressInput struct {
	ID            uuid.UUID
	Label         string
	RecipientName string
	Phone         string
	Country       string
	City          string
	Street        string
	Postcode      string
	Latitude      float64
	Longitude     float64
	IsDefault     bool
}

func (s *Service) Create(ctx context.Context, in addressInput) (domain.Address, error) {
	addr, err := s.repo.Create(ctx, domain.Address{
		ID:            in.ID,
		UserID:        uuid.UUID{},
		Label:         in.Label,
		RecipientName: in.RecipientName,
		Phone:         in.Phone,
		Country:       in.Country,
		Province:      "",
		City:          in.City,
		Street:        in.Street,
		PostalCode:    "",
		Lat:           0,
		Lng:           0,
		IsDefault:     in.IsDefault,
		CreatedAt:     time.Time{},
		UpdatedAt:     time.Time{},
	})
	if err != nil {
		return domain.Address{}, err
	}
	return addr, nil
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (domain.Address, error) {
	addr, err := s.repo.GetByID(ctx, id, userID)
	if err != nil {
		return domain.Address{}, err
	}
	return addr, nil
}
func (s *Service) getAllAddresses(ctx context.Context, userID uuid.UUID) ([]domain.Address, error) {
	addresses, err := s.repo.GetAllAddresses(ctx, userID)
	if err != nil {
		return []domain.Address{}, err
	}
	return addresses, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (int64, error) {
	addr, err := s.repo.Delete(ctx, id, userID)
	if err != nil {
		return -1, err
	}
	return addr, err

}

func (s *Service) update(ctx context.Context, u domain.Address) (domain.Address, error) {
	addr, err := s.repo.update(ctx, u)
	if err != nil {
		return domain.Address{}, err
	}
	return addr, nil
}
