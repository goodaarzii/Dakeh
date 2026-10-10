package variant

import (
	"time"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// variantRequest is used for both create (POST) and full update (PUT).
// Price and CompareAtPrice are in USD cents. CompareAtPrice 0 means "no old price".
// Price is a pointer so that a missing price is rejected, but a real 0 is accepted.
// "CompareAtPrice must be greater than Price" is checked in the usecase.
type variantRequest struct {
	SKU            string            `json:"sku" binding:"required,max=64"`
	Name           string            `json:"name" binding:"max=255"`
	Price          *int64            `json:"price" binding:"required,gte=0"`
	CompareAtPrice int64             `json:"compare_at_price" binding:"gte=0"`
	Attributes     map[string]string `json:"attributes" binding:"max=20,dive,keys,max=50,endkeys,max=100"`
	WeightGrams    int               `json:"weight_grams" binding:"gte=0"`
	IsActive       *bool             `json:"is_active"`
}

// createVariantRequest adds the starting stock. The fields of variantRequest stay at the top level in JSON.
type createVariantRequest struct {
	variantRequest
	InitialStock      int `json:"initial_stock" binding:"gte=0"`
	LowStockThreshold int `json:"low_stock_threshold" binding:"gte=0"`
}

// toDomain leaves ID and ProductID empty: the handler sets them from the URL.
// Attributes is never nil, because a nil map becomes JSON null and the database CHECK refuses it.
func (r variantRequest) toDomain() domain.Variant {
	active := true
	if r.IsActive != nil {
		active = *r.IsActive
	}
	attrs := r.Attributes
	if attrs == nil {
		attrs = map[string]string{}
	}
	return domain.Variant{
		SKU:            r.SKU,
		Name:           r.Name,
		Price:          *r.Price, // safe: binding "required" already rejected nil
		CompareAtPrice: r.CompareAtPrice,
		Attributes:     attrs,
		WeightGrams:    r.WeightGrams,
		IsActive:       active,
	}
}

// setStockRequest sets the absolute quantity. QuantityOnHand is a pointer so a real 0 is accepted.
type setStockRequest struct {
	QuantityOnHand    *int `json:"quantity_on_hand" binding:"required,gte=0"`
	LowStockThreshold int  `json:"low_stock_threshold" binding:"gte=0"`
}

// adjustStockRequest adds or removes stock: +50 restock, -3 write-off. A delta of 0 is rejected as pointless.
type adjustStockRequest struct {
	Delta int `json:"delta" binding:"required,min=-1000000,max=1000000"`
}

// stockResponse is the admin view of stock. Customers will later get a smaller one (in_stock only).
type stockResponse struct {
	QuantityOnHand    int `json:"quantity_on_hand"`
	QuantityReserved  int `json:"quantity_reserved"`
	Available         int `json:"available"`
	LowStockThreshold int `json:"low_stock_threshold"`
}

type variantResponse struct {
	ID             uuid.UUID         `json:"id"`
	ProductID      uuid.UUID         `json:"product_id"`
	SKU            string            `json:"sku"`
	Name           string            `json:"name"`
	Price          int64             `json:"price"`            // cents
	CompareAtPrice int64             `json:"compare_at_price"` // cents, 0 = none
	Attributes     map[string]string `json:"attributes"`
	WeightGrams    int               `json:"weight_grams"`
	IsActive       bool              `json:"is_active"`
	Stock          stockResponse     `json:"stock"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

func toStockResponse(i domain.Inventory) stockResponse {
	return stockResponse{
		QuantityOnHand:    i.QuantityOnHand,
		QuantityReserved:  i.QuantityReserved,
		Available:         i.Available(),
		LowStockThreshold: i.LowStockThreshold,
	}
}

func toVariantResponse(v domain.VariantWithStock) variantResponse {
	attrs := v.Variant.Attributes
	if attrs == nil {
		attrs = map[string]string{} // {} and never null
	}
	return variantResponse{
		ID:             v.Variant.ID,
		ProductID:      v.Variant.ProductID,
		SKU:            v.Variant.SKU,
		Name:           v.Variant.Name,
		Price:          v.Variant.Price,
		CompareAtPrice: v.Variant.CompareAtPrice,
		Attributes:     attrs,
		WeightGrams:    v.Variant.WeightGrams,
		IsActive:       v.Variant.IsActive,
		Stock:          toStockResponse(v.Stock),
		CreatedAt:      v.Variant.CreatedAt,
		UpdatedAt:      v.Variant.UpdatedAt,
	}
}

func toVariantResponses(list []domain.VariantWithStock) []variantResponse {
	out := make([]variantResponse, len(list))
	for i, v := range list {
		out[i] = toVariantResponse(v)
	}
	return out
}

// listStockRequest: the variant IDs to look up. Limited to 100 per call.
type listStockRequest struct {
	VariantIDs []uuid.UUID `json:"variant_ids" binding:"required,min=1,max=100,dive,required"`
}

// variantStockResponse is a stock together with the variant it belongs to.
type variantStockResponse struct {
	VariantID uuid.UUID `json:"variant_id"`
	stockResponse
}

func toVariantStockResponses(list []domain.Inventory) []variantStockResponse {
	out := make([]variantStockResponse, len(list))
	for i, inv := range list {
		out[i] = variantStockResponse{VariantID: inv.VariantID, stockResponse: toStockResponse(inv)}
	}
	return out
}
