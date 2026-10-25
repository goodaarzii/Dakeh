package order

import (
	"context"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// Shipping is currently a flat-rate demo policy. Keep the frontend display in
// CheckoutPage.tsx in sync if this policy changes. Amounts are USD cents.
const DefaultShippingFeeCents int64 = 500

const (
	defaultPageSize = 10
	maxPageSize     = 50
	maxPage         = 1_000_000
)

type CartStore interface {
	Get(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error)
	Clear(ctx context.Context, userID uuid.UUID) error
}

type Repository interface {
	Create(ctx context.Context, userID, addressID uuid.UUID, idempotencyKey string, cartItems []domain.CartItem, shippingFee int64) (domain.Order, error)
	FindByIdempotencyKey(ctx context.Context, userID uuid.UUID, key string) (domain.Order, bool, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Order, int, error)
	Cancel(ctx context.Context, userID, orderID uuid.UUID) (domain.Order, error)
}

type Service struct {
	cart CartStore
	repo Repository
}

func NewService(cart CartStore, repo Repository) *Service {
	return &Service{cart: cart, repo: repo}
}

// Create submits the current account cart. A PostgreSQL transaction creates the
// order and reserves inventory atomically. Payment is not collected in this phase.
// Cart clearing is best-effort after commit: a Redis failure must not turn a
// successfully-created order into an API error that might cause duplicate retries.
func (s *Service) Create(ctx context.Context, userID, addressID uuid.UUID, idempotencyKey string) (domain.Order, bool, error) {
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if userID == uuid.Nil || addressID == uuid.Nil || idempotencyKey == "" || len(idempotencyKey) > 128 {
		return domain.Order{}, false, domain.ErrInvalidInput
	}

	// A retry after a successful submit returns the same order, even if the original
	// request already cleared the Redis cart.
	if existing, found, err := s.repo.FindByIdempotencyKey(ctx, userID, idempotencyKey); err != nil {
		return domain.Order{}, false, err
	} else if found {
		return existing, s.clearCart(ctx, userID, existing.ID), nil
	}

	items, err := s.cart.Get(ctx, userID)
	if err != nil {
		return domain.Order{}, false, err
	}
	if len(items) == 0 {
		return domain.Order{}, false, domain.ErrInvalidInput
	}

	created, err := s.repo.Create(ctx, userID, addressID, idempotencyKey, items, DefaultShippingFeeCents)
	if err != nil {
		// Same-key requests can race before the unique constraint is reached: the second
		// may observe stock already reserved by the first. Resolve any concurrent winner
		// by key before reporting the original error.
		if existing, found, findErr := s.repo.FindByIdempotencyKey(ctx, userID, idempotencyKey); findErr == nil && found {
			return existing, s.clearCart(ctx, userID, existing.ID), nil
		}
		return domain.Order{}, false, err
	}
	return created, s.clearCart(ctx, userID, created.ID), nil
}

func (s *Service) clearCart(ctx context.Context, userID, orderID uuid.UUID) bool {
	if err := s.cart.Clear(ctx, userID); err != nil {
		slog.Error("order was submitted but cart could not be cleared", "order_id", orderID, "user_id", userID, "err", err)
		return false
	}
	return true
}

func (s *Service) ListMine(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]domain.Order, int, bool, error) {
	if userID == uuid.Nil {
		return nil, 0, false, domain.ErrInvalidInput
	}
	if page < 1 {
		page = 1
	}
	if page > maxPage {
		return nil, 0, false, domain.ErrInvalidInput
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	items, total, err := s.repo.ListByUser(ctx, userID, pageSize+1, (page-1)*pageSize)
	if err != nil {
		return nil, 0, false, err
	}
	hasMore := len(items) > pageSize
	if hasMore {
		items = items[:pageSize]
	}
	return items, total, hasMore, nil
}

func (s *Service) CancelMine(ctx context.Context, userID, orderID uuid.UUID) (domain.Order, error) {
	if userID == uuid.Nil || orderID == uuid.Nil {
		return domain.Order{}, domain.ErrInvalidInput
	}
	return s.repo.Cancel(ctx, userID, orderID)
}
