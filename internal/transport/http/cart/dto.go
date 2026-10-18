package cart

import (
	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// setQuantityRequest: PUT /cart/items/:variant_id
// The "max=20" below must match domain.MaxQuantityPerLine (a tag cannot use a constant).
type setQuantityRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1,max=20"`
}

// mergeRequest: POST /cart/merge, the guest cart that the browser kept until login.
type mergeRequest struct {
	Items []mergeItem `json:"items" binding:"required,min=1,max=50,dive"`
}

type mergeItem struct {
	VariantID uuid.UUID `json:"variant_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,min=1,max=20"`
}

func (r mergeRequest) toDomain() []domain.CartItem {
	items := make([]domain.CartItem, len(r.Items))
	for i, it := range r.Items {
		items[i] = domain.CartItem{VariantID: it.VariantID, Quantity: it.Quantity}
	}
	return items
}

// cartLineResponse: prices are USD cents. The product fields are left out for a line that is not available any more.
type cartLineResponse struct {
	VariantID      uuid.UUID         `json:"variant_id"`
	Quantity       int               `json:"quantity"`
	Available      bool              `json:"available"` // false: no longer sold, the customer should remove it
	InStock        bool              `json:"in_stock"`  // false: not enough stock for this quantity
	ProductID      *uuid.UUID        `json:"product_id,omitempty"`
	ProductName    string            `json:"product_name,omitempty"`
	ProductSlug    string            `json:"product_slug,omitempty"`
	VariantName    string            `json:"variant_name,omitempty"`
	SKU            string            `json:"sku,omitempty"`
	Attributes     map[string]string `json:"attributes,omitempty"`
	ImageURL       string            `json:"image_url,omitempty"`
	UnitPrice      int64             `json:"unit_price"`
	CompareAtPrice int64             `json:"compare_at_price"`
	LineTotal      int64             `json:"line_total"`
}

type cartResponse struct {
	Items       []cartLineResponse `json:"items"`
	Subtotal    int64              `json:"subtotal"` // cents, shipping is added at checkout
	ItemCount   int                `json:"item_count"`
	CanCheckout bool               `json:"can_checkout"`
}

func toCartResponse(c domain.Cart) cartResponse {
	items := make([]cartLineResponse, len(c.Lines))
	for i, l := range c.Lines {
		line := cartLineResponse{
			VariantID: l.VariantID,
			Quantity:  l.Quantity,
			Available: l.Sellable,
			InStock:   l.InStock,
			LineTotal: l.LineTotal,
		}
		if l.Sellable {
			v := l.Variant
			productID := v.ProductID
			line.ProductID = &productID
			line.ProductName = v.ProductName
			line.ProductSlug = v.ProductSlug
			line.VariantName = v.VariantName
			line.SKU = v.SKU
			line.Attributes = v.Attributes
			line.ImageURL = v.ImageURL
			line.UnitPrice = v.UnitPrice
			line.CompareAtPrice = v.CompareAtPrice
		}
		items[i] = line
	}
	return cartResponse{
		Items:       items,
		Subtotal:    c.Subtotal,
		ItemCount:   c.ItemCount,
		CanCheckout: c.CanCheckout,
	}
}
