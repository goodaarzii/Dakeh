package variant

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// Repository is what this usecase needs from storage (implemented by postgres.VariantRepo).
type Repository interface {
	Create(ctx context.Context, v domain.Variant, initialStock, lowStockThreshold int) (domain.VariantWithStock, error)
	Update(ctx context.Context, v domain.Variant) (domain.Variant, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Variant, error) // Variant, not VariantWithStock
	ListByProduct(ctx context.Context, productID uuid.UUID) ([]domain.Variant, error)
	Delete(ctx context.Context, id uuid.UUID) error

	GetStock(ctx context.Context, variantID uuid.UUID) (domain.Inventory, error)
	ListStock(ctx context.Context, variantIDs []uuid.UUID) ([]domain.Inventory, error)
	SetStock(ctx context.Context, variantID uuid.UUID, onHand, lowStockThreshold int) (domain.Inventory, error)
	AdjustStock(ctx context.Context, variantID uuid.UUID, delta int) (domain.Inventory, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create returns the variant together with its stock (the repo already has both).
func (s *Service) Create(ctx context.Context, v domain.Variant, initialStock, lowStockThreshold int) (domain.VariantWithStock, error) {
	if err := validatePrices(v); err != nil {
		return domain.VariantWithStock{}, err
	}
	return s.repo.Create(ctx, v, initialStock, lowStockThreshold)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (domain.VariantWithStock, error) {
	v, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.VariantWithStock{}, err
	}
	stock, err := s.repo.GetStock(ctx, id)
	if err != nil {
		return domain.VariantWithStock{}, err
	}
	return domain.VariantWithStock{Variant: v, Stock: stock}, nil
}

// Update changes the variant fields (not the stock), then loads the stock so the response is complete.
func (s *Service) Update(ctx context.Context, v domain.Variant) (domain.VariantWithStock, error) {
	if err := validatePrices(v); err != nil {
		return domain.VariantWithStock{}, err
	}
	updated, err := s.repo.Update(ctx, v)
	if err != nil {
		return domain.VariantWithStock{}, err
	}
	stock, err := s.repo.GetStock(ctx, updated.ID)
	if err != nil {
		return domain.VariantWithStock{}, err
	}
	return domain.VariantWithStock{Variant: updated, Stock: stock}, nil
}

// ListByProduct returns all variants of a product with their stock: 2 queries in total, however many variants.
func (s *Service) ListByProduct(ctx context.Context, productID uuid.UUID) ([]domain.VariantWithStock, error) {
	variants, err := s.repo.ListByProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, len(variants))
	for i, v := range variants {
		ids[i] = v.ID
	}
	stocks, err := s.repo.ListStock(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Look up the stock of each variant by its ID.
	// A variant with no stock row gets the zero value (available = 0), which is the safe answer.
	byVariant := make(map[uuid.UUID]domain.Inventory, len(stocks))
	for _, st := range stocks {
		byVariant[st.VariantID] = st
	}

	out := make([]domain.VariantWithStock, len(variants))
	for i, v := range variants {
		out[i] = domain.VariantWithStock{Variant: v, Stock: byVariant[v.ID]}
	}
	return out, nil
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) SetStock(ctx context.Context, variantID uuid.UUID, onHand, lowStockThreshold int) (domain.Inventory, error) {
	return s.repo.SetStock(ctx, variantID, onHand, lowStockThreshold)
}

func (s *Service) AdjustStock(ctx context.Context, variantID uuid.UUID, delta int) (domain.Inventory, error) {
	return s.repo.AdjustStock(ctx, variantID, delta)
}

func (s *Service) GetStock(ctx context.Context, variantID uuid.UUID) (domain.Inventory, error) {
	return s.repo.GetStock(ctx, variantID)
}

func (s *Service) ListStock(ctx context.Context, variantIDs []uuid.UUID) ([]domain.Inventory, error) {
	return s.repo.ListStock(ctx, variantIDs)
}

func validatePrices(v domain.Variant) error {
	if v.CompareAtPrice != 0 && v.CompareAtPrice <= v.Price {
		return fmt.Errorf("%w: compare_at_price must be greater than price", domain.ErrInvalidInput)
	}
	return nil
}
