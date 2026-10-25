package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	OrderStatusSubmitted = "submitted"
	OrderStatusCancelled = "cancelled"
)

// OrderAddress is a delivery-address snapshot. Later edits to a saved address must
// not change where an existing order was intended to go.
type OrderAddress struct {
	Label         string
	RecipientName string
	Phone         string
	Country       string
	Province      string
	City          string
	Street        string
	PostalCode    string
}

// OrderItem is a catalogue snapshot taken when the order is submitted.
type OrderItem struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	ProductID   uuid.UUID
	VariantID   uuid.UUID
	ProductName string
	ProductSlug string
	VariantName string
	SKU         string
	Attributes  map[string]string
	ImageURL    string
	UnitPrice   int64
	Quantity    int
	LineTotal   int64
	CreatedAt   time.Time
}

type Order struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Status      string
	Currency    string
	Subtotal    int64
	ShippingFee int64
	Total       int64
	Address     OrderAddress
	Items       []OrderItem
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CancelledAt *time.Time
}
