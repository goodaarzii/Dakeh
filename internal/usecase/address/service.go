package address

import (
	"context"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, address domain.Address) (domain.Address, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (domain.Address, error)
	GetUserAddresses(ctx context.Context, userID uuid.UUID) ([]domain.Address, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (int64, error)
	Update(ctx context.Context, u domain.Address) (domain.Address, error)
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

func (s *Service) Create(ctx context.Context, in domain.Address) (domain.Address, error) {
	addr, err := s.repo.Create(ctx, domain.Address{
		ID:            in.ID,
		UserID:        in.UserID,
		Label:         in.Label,
		RecipientName: in.RecipientName,
		Phone:         in.Phone,
		Country:       in.Country,
		Province:      in.Province,
		City:          in.City,
		Street:        in.Street,
		PostalCode:    in.PostalCode,
		Lat:           in.Lat,
		Lng:           in.Lng,
		IsDefault:     in.IsDefault,
		CreatedAt:     in.UpdatedAt,
		UpdatedAt:     in.CreatedAt,
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
func (s *Service) GetUserAddresses(ctx context.Context, userID uuid.UUID) ([]domain.Address, error) {
	addresses, err := s.repo.GetUserAddresses(ctx, userID)
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

func (s *Service) Update(ctx context.Context, u domain.Address) (domain.Address, error) {
	addr, err := s.repo.Update(ctx, u)
	if err != nil {
		return domain.Address{}, err
	}
	return addr, nil
}
