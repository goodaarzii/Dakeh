package domain

import (
	"time"

	"github.com/google/uuid"
)

type ProductStatus string

const (
	ProductStatusDraft    ProductStatus = "draft"    // not visible to customers
	ProductStatusActive   ProductStatus = "active"   // visible and sellable
	ProductStatusArchived ProductStatus = "archived" // hidden, kept for old orders
)

// Valid reports whether s is one of the known statuses.
func (s ProductStatus) Valid() bool {
	switch s {
	case ProductStatusDraft, ProductStatusActive, ProductStatusArchived:
		return true
	}
	return false
}

// Product is the item we sell. Price and stock live on its variants, not here.
// Soft-deleted products never reach the domain, so there is no DeletedAt.
type Product struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	Name        string
	Slug        string
	Description string
	Brand       string
	Status      ProductStatus
	AvgRating   float64
	RatingCount int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
