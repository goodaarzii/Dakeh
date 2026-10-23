package favorite

import (
	"context"
	"testing"
	"time"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

type fakeRepository struct {
	added       domain.Favorite
	removedUser uuid.UUID
	removedProd uuid.UUID
	listedUser  uuid.UUID
	addCalls    int
	removeCalls int
	listCalls   int
}

func (r *fakeRepository) AddFavorite(_ context.Context, favorite domain.Favorite) (domain.Favorite, error) {
	r.addCalls++
	r.added = favorite
	favorite.CreatedAt = time.Date(2026, time.October, 23, 0, 0, 0, 0, time.UTC)
	return favorite, nil
}

func (r *fakeRepository) RemoveFavorite(_ context.Context, userID, productID uuid.UUID) error {
	r.removeCalls++
	r.removedUser = userID
	r.removedProd = productID
	return nil
}

func (r *fakeRepository) ListFavoriteProducts(_ context.Context, userID uuid.UUID) ([]domain.FavoriteProduct, error) {
	r.listCalls++
	r.listedUser = userID
	return []domain.FavoriteProduct{}, nil
}

func TestAddPassesTheAuthenticatedUserAndProductToRepository(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)
	userID, productID := uuid.New(), uuid.New()

	favorite, err := svc.Add(context.Background(), userID, productID)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if repo.addCalls != 1 || repo.added.UserID != userID || repo.added.ProductID != productID {
		t.Fatalf("Add() did not persist the supplied IDs: %#v", repo)
	}
	if favorite.CreatedAt.IsZero() {
		t.Fatal("Add() should return the saved favorite timestamp")
	}
}

func TestUsecaseRejectsNilIDsBeforeRepositoryAccess(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)

	if _, err := svc.Add(context.Background(), uuid.Nil, uuid.New()); err == nil {
		t.Fatal("Add() should reject a nil user ID")
	}
	if _, err := svc.Add(context.Background(), uuid.New(), uuid.Nil); err == nil {
		t.Fatal("Add() should reject a nil product ID")
	}
	if err := svc.Remove(context.Background(), uuid.Nil, uuid.New()); err == nil {
		t.Fatal("Remove() should reject a nil user ID")
	}
	if _, err := svc.List(context.Background(), uuid.Nil); err == nil {
		t.Fatal("List() should reject a nil user ID")
	}
	if repo.addCalls != 0 || repo.removeCalls != 0 || repo.listCalls != 0 {
		t.Fatal("invalid IDs should be rejected before repository access")
	}
}
