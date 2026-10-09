package domain

import (
	"time"

	"github.com/google/uuid"
)

// Variant is the sellable version of a product (a size, a color, ...).
// Money is stored in integer USD cents: 1999 means $19.99.
// Soft-deleted variants never reach the domain, so there is no DeletedAt.
type Variant struct {
	ID             uuid.UUID
	ProductID      uuid.UUID
	SKU            string
	Name           string
	Price          int64 // cents
	CompareAtPrice int64 // old price in cents, 0 = none
	Attributes     map[string]string
	WeightGrams    int
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Inventory is the stock of one variant.
type Inventory struct {
	VariantID         uuid.UUID
	QuantityOnHand    int
	QuantityReserved  int // held by unpaid orders
	LowStockThreshold int
	UpdatedAt         time.Time
}

// Available is what a customer can still buy.
func (i Inventory) Available() int {
	return i.QuantityOnHand - i.QuantityReserved
}

// VariantWithStock is what most screens need: a variant and its stock together.
type VariantWithStock struct {
	Variant Variant
	Stock   Inventory
}
