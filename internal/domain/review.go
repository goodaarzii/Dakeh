package domain

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ReviewDraft contains the customer-editable fields. Product and user IDs are
// supplied separately from the URL and authenticated access token.
type ReviewDraft struct {
	Rating int
	Title  string
	Body   string
}

// Review is one customer's review for a product.
type Review struct {
	ID           uuid.UUID
	ProductID    uuid.UUID
	UserID       uuid.UUID
	ReviewerName string
	Rating       int
	Title        string
	Body         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewReview validates identity and customer input before the repository is called.
func NewReview(userID, productID uuid.UUID, rating int, title, body string) (Review, error) {
	if userID == uuid.Nil || productID == uuid.Nil {
		return Review{}, ErrInvalidInput
	}
	draft, err := NewReviewDraft(rating, title, body)
	if err != nil {
		return Review{}, err
	}
	return Review{
		UserID: userID, ProductID: productID,
		Rating: draft.Rating, Title: draft.Title, Body: draft.Body,
	}, nil
}

// NewReviewDraft normalizes text and enforces the rules independently of HTTP validation.
func NewReviewDraft(rating int, title, body string) (ReviewDraft, error) {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)

	if rating < 1 || rating > 5 || utf8.RuneCountInString(title) > 120 {
		return ReviewDraft{}, ErrInvalidInput
	}
	bodyLength := utf8.RuneCountInString(body)
	if bodyLength < 3 || bodyLength > 5000 {
		return ReviewDraft{}, ErrInvalidInput
	}
	return ReviewDraft{Rating: rating, Title: title, Body: body}, nil
}

// BelongsTo verifies ownership before exposing or changing a review.
func (r Review) BelongsTo(userID uuid.UUID) bool {
	return userID != uuid.Nil && r.UserID == userID
}
