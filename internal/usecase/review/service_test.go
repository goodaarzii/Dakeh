package review

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

type fakeRepository struct {
	listLimit   int
	listOffset  int
	created     domain.Review
	createCalls int
}

func (f *fakeRepository) ListByProductID(_ context.Context, _ uuid.UUID, limit, offset int) ([]domain.Review, int, error) {
	f.listLimit, f.listOffset = limit, offset
	items := make([]domain.Review, limit)
	for i := range items {
		items[i] = domain.Review{ID: uuid.New()}
	}
	return items, 25, nil
}
func (f *fakeRepository) GetMineForProduct(context.Context, uuid.UUID, uuid.UUID) (domain.Review, error) {
	return domain.Review{}, nil
}
func (f *fakeRepository) Create(_ context.Context, review domain.Review) (domain.Review, error) {
	f.createCalls++
	f.created = review
	return review, nil
}
func (f *fakeRepository) UpdateMine(context.Context, uuid.UUID, uuid.UUID, domain.ReviewDraft) (domain.Review, error) {
	return domain.Review{}, nil
}
func (f *fakeRepository) DeleteMine(context.Context, uuid.UUID, uuid.UUID) error { return nil }

func TestListUsesExtraRowForHasMore(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)
	items, count, hasMore, err := svc.ListByProductID(context.Background(), uuid.New(), 2, 10)
	if err != nil {
		t.Fatal(err)
	}
	if count != 25 || !hasMore || len(items) != 10 {
		t.Fatalf("unexpected page: count=%d hasMore=%v len=%d", count, hasMore, len(items))
	}
	if repo.listLimit != 11 || repo.listOffset != 10 {
		t.Fatalf("query limit/offset = %d/%d", repo.listLimit, repo.listOffset)
	}
}

func TestCreatePassesAuthenticatedUserAndValidatedDraft(t *testing.T) {
	repo := &fakeRepository{}
	svc := NewService(repo)
	userID, productID := uuid.New(), uuid.New()
	_, err := svc.Create(context.Background(), userID, productID, 5, " Great ", " Nice fit ")
	if err != nil {
		t.Fatal(err)
	}
	if repo.createCalls != 1 || repo.created.UserID != userID || repo.created.ProductID != productID || repo.created.Title != "Great" || repo.created.Body != "Nice fit" {
		t.Fatalf("unexpected repository input: %#v", repo)
	}
	if _, err := svc.Create(context.Background(), userID, productID, 9, "", "valid body"); err == nil {
		t.Fatal("expected invalid rating to be rejected before repository access")
	}
	if repo.createCalls != 1 {
		t.Fatal("invalid draft should not reach repository")
	}
}
