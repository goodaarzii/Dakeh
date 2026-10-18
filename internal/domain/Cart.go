package domain

import "github.com/google/uuid"

// Limits of a cart (business rules).
const (
	MaxCartLines       = 50 // different variants in one cart
	MaxQuantityPerLine = 20 // keep in step with the "max=20" validation tags in the cart DTO
)

// CartItem is all that Redis stores: which variant, and how many. NO price: prices always come from Postgres.
type CartItem struct {
	VariantID uuid.UUID
	Quantity  int
}

// SellableVariant is the live data of a variant that can be bought right now.
type SellableVariant struct {
	VariantID      uuid.UUID
	ProductID      uuid.UUID
	ProductName    string
	ProductSlug    string
	VariantName    string
	SKU            string
	Attributes     map[string]string
	ImageURL       string // "" when the product has no image
	UnitPrice      int64  // cents
	CompareAtPrice int64  // cents, 0 = none
	Available      int    // units that can still be bought
}

// CartLine is a cart item together with its live data.
type CartLine struct {
	VariantID uuid.UUID
	Quantity  int
	Sellable  bool            // false: the variant was deleted, deactivated, or its product is hidden
	InStock   bool            // true: enough stock for Quantity
	Variant   SellableVariant // zero value when Sellable is false
	LineTotal int64           // UnitPrice * Quantity, cents
}

// Cart is what the cart page shows.
type Cart struct {
	Lines       []CartLine
	Subtotal    int64 // cents, sellable lines only. Shipping is added at checkout.
	ItemCount   int   // total quantity
	CanCheckout bool  // not empty, and every line is sellable and in stock
}
