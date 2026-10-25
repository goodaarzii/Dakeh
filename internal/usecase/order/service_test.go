package order

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

type fakeCartStore struct {
	items      []domain.CartItem
	getCalls   int
	clearCalls int
	getErr     error
	clearErr   error
}

func (f *fakeCartStore) Get(context.Context, uuid.UUID) ([]domain.CartItem, error) {
	f.getCalls++
	return append([]domain.CartItem(nil), f.items...), f.getErr
}

func (f *fakeCartStore) Clear(context.Context, uuid.UUID) error {
	f.clearCalls++
	return f.clearErr
}

type fakeOrderRepository struct {
	createdOrder  domain.Order
	existingOrder domain.Order
	found         bool
	findErr       error
	createErr     error
	createCalls   int
	gotUserID     uuid.UUID
	gotAddressID  uuid.UUID
	gotKey        string
	gotItems      []domain.CartItem
	gotShipping   int64
}

func (f *fakeOrderRepository) Create(_ context.Context, userID, addressID uuid.UUID, key string, items []domain.CartItem, fee int64) (domain.Order, error) {
	f.createCalls++
	f.gotUserID = userID
	f.gotAddressID = addressID
	f.gotKey = key
	f.gotItems = append([]domain.CartItem(nil), items...)
	f.gotShipping = fee
	return f.createdOrder, f.createErr
}

func (f *fakeOrderRepository) FindByIdempotencyKey(context.Context, uuid.UUID, string) (domain.Order, bool, error) {
	return f.existingOrder, f.found, f.findErr
}

func (f *fakeOrderRepository) ListByUser(context.Context, uuid.UUID, int, int) ([]domain.Order, int, error) {
	return []domain.Order{}, 0, nil
}

func (f *fakeOrderRepository) Cancel(context.Context, uuid.UUID, uuid.UUID) (domain.Order, error) {
	return domain.Order{}, nil
}

func TestCreateSubmitsCartAndClearsIt(t *testing.T) {
	userID, addressID, variantID := uuid.New(), uuid.New(), uuid.New()
	cart := &fakeCartStore{items: []domain.CartItem{{VariantID: variantID, Quantity: 2}}}
	repo := &fakeOrderRepository{createdOrder: domain.Order{ID: uuid.New(), UserID: userID, Status: domain.OrderStatusSubmitted}}
	svc := NewService(cart, repo)

	created, cleared, err := svc.Create(context.Background(), userID, addressID, "checkout-123")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if !cleared || cart.getCalls != 1 || cart.clearCalls != 1 {
		t.Fatalf("expected cart to be read and cleared, cleared=%v getCalls=%d clearCalls=%d", cleared, cart.getCalls, cart.clearCalls)
	}
	if repo.createCalls != 1 || repo.gotUserID != userID || repo.gotAddressID != addressID || repo.gotKey != "checkout-123" {
		t.Fatalf("unexpected repository arguments: %#v", repo)
	}
	if len(repo.gotItems) != 1 || repo.gotItems[0].VariantID != variantID || repo.gotItems[0].Quantity != 2 {
		t.Fatalf("repository received wrong cart items: %#v", repo.gotItems)
	}
	if repo.gotShipping != DefaultShippingFeeCents {
		t.Fatalf("shipping fee = %d, want %d", repo.gotShipping, DefaultShippingFeeCents)
	}
	if created.Status != domain.OrderStatusSubmitted {
		t.Fatalf("status = %q, want %q", created.Status, domain.OrderStatusSubmitted)
	}
}

func TestCreateRetryReturnsExistingOrderWithoutReadingCart(t *testing.T) {
	userID, addressID := uuid.New(), uuid.New()
	existing := domain.Order{ID: uuid.New(), UserID: userID, Status: domain.OrderStatusSubmitted}
	cart := &fakeCartStore{}
	repo := &fakeOrderRepository{existingOrder: existing, found: true}
	svc := NewService(cart, repo)

	got, cleared, err := svc.Create(context.Background(), userID, addressID, "same-key")
	if err != nil {
		t.Fatalf("Create() retry error = %v", err)
	}
	if got.ID != existing.ID || !cleared || cart.getCalls != 0 || repo.createCalls != 0 {
		t.Fatalf("retry created a duplicate or failed to clear cart: order=%s cleared=%v cartGets=%d creates=%d", got.ID, cleared, cart.getCalls, repo.createCalls)
	}
}

func TestCreateReturnsOrderWhenRedisCartCannotBeCleared(t *testing.T) {
	userID, addressID := uuid.New(), uuid.New()
	created := domain.Order{ID: uuid.New(), UserID: userID, Status: domain.OrderStatusSubmitted}
	cart := &fakeCartStore{items: []domain.CartItem{{VariantID: uuid.New(), Quantity: 1}}, clearErr: errors.New("redis unavailable")}
	repo := &fakeOrderRepository{createdOrder: created}
	svc := NewService(cart, repo)

	got, cleared, err := svc.Create(context.Background(), userID, addressID, "retry-key")
	if err != nil {
		t.Fatalf("Create() should not report failure after the DB commit: %v", err)
	}
	if got.ID != created.ID || cleared {
		t.Fatalf("got order=%s, cartCleared=%v; want order=%s and false", got.ID, cleared, created.ID)
	}
}

func TestCreateRejectsInvalidInputsBeforeRepositoryAccess(t *testing.T) {
	cart := &fakeCartStore{items: []domain.CartItem{{VariantID: uuid.New(), Quantity: 1}}}
	repo := &fakeOrderRepository{}
	svc := NewService(cart, repo)

	if _, _, err := svc.Create(context.Background(), uuid.Nil, uuid.New(), "key"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("nil user should be invalid input, got %v", err)
	}
	if _, _, err := svc.Create(context.Background(), uuid.New(), uuid.Nil, "key"); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("nil address should be invalid input, got %v", err)
	}
	if _, _, err := svc.Create(context.Background(), uuid.New(), uuid.New(), "  "); !errors.Is(err, domain.ErrInvalidInput) {
		t.Fatalf("empty idempotency key should be invalid input, got %v", err)
	}
	if repo.createCalls != 0 || cart.getCalls != 0 {
		t.Fatal("invalid input should be rejected before touching the cart or repository")
	}
}
