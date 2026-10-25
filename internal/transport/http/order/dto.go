package order

import (
	"time"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

type createOrderRequest struct {
	AddressID uuid.UUID `json:"address_id" binding:"required"`
}

type orderAddressResponse struct {
	Label         string `json:"label"`
	RecipientName string `json:"recipient_name"`
	Phone         string `json:"phone"`
	Country       string `json:"country"`
	Province      string `json:"province"`
	City          string `json:"city"`
	Street        string `json:"street"`
	PostalCode    string `json:"postal_code"`
}

type orderItemResponse struct {
	ID          uuid.UUID         `json:"id"`
	ProductID   uuid.UUID         `json:"product_id"`
	VariantID   uuid.UUID         `json:"variant_id"`
	ProductName string            `json:"product_name"`
	ProductSlug string            `json:"product_slug"`
	VariantName string            `json:"variant_name"`
	SKU         string            `json:"sku"`
	Attributes  map[string]string `json:"attributes"`
	ImageURL    string            `json:"image_url"`
	UnitPrice   int64             `json:"unit_price"`
	Quantity    int               `json:"quantity"`
	LineTotal   int64             `json:"line_total"`
}

type orderResponse struct {
	ID          uuid.UUID            `json:"id"`
	Status      string               `json:"status"`
	Currency    string               `json:"currency"`
	Subtotal    int64                `json:"subtotal"`
	ShippingFee int64                `json:"shipping_fee"`
	Total       int64                `json:"total"`
	Address     orderAddressResponse `json:"shipping_address"`
	Items       []orderItemResponse  `json:"items"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	CancelledAt *time.Time           `json:"cancelled_at,omitempty"`
}

type createOrderResponse struct {
	Order       orderResponse `json:"order"`
	CartCleared bool          `json:"cart_cleared"`
}

type orderListResponse struct {
	Items    []orderResponse `json:"items"`
	Count    int             `json:"count"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
	HasMore  bool            `json:"has_more"`
}

type listQuery struct {
	Page     int `form:"page" binding:"omitempty,gte=1"`
	PageSize int `form:"page_size" binding:"omitempty,gte=1,lte=50"`
}

func toOrderResponse(item domain.Order) orderResponse {
	items := make([]orderItemResponse, len(item.Items))
	for i, line := range item.Items {
		attrs := line.Attributes
		if attrs == nil {
			attrs = map[string]string{}
		}
		items[i] = orderItemResponse{
			ID: line.ID, ProductID: line.ProductID, VariantID: line.VariantID,
			ProductName: line.ProductName, ProductSlug: line.ProductSlug,
			VariantName: line.VariantName, SKU: line.SKU, Attributes: attrs,
			ImageURL: line.ImageURL, UnitPrice: line.UnitPrice, Quantity: line.Quantity,
			LineTotal: line.LineTotal,
		}
	}
	if items == nil {
		items = []orderItemResponse{}
	}
	return orderResponse{
		ID: item.ID, Status: item.Status, Currency: item.Currency,
		Subtotal: item.Subtotal, ShippingFee: item.ShippingFee, Total: item.Total,
		Address: orderAddressResponse{
			Label: item.Address.Label, RecipientName: item.Address.RecipientName,
			Phone: item.Address.Phone, Country: item.Address.Country,
			Province: item.Address.Province, City: item.Address.City,
			Street: item.Address.Street, PostalCode: item.Address.PostalCode,
		},
		Items: items, CreatedAt: item.CreatedAt, UpdatedAt: item.UpdatedAt,
		CancelledAt: item.CancelledAt,
	}
}
