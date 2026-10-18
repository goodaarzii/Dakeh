package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// ListSellableVariants returns the live data of the variants that can be bought right now.
// Variants that are not sellable are simply missing from the result.
// It is a method of CatalogRepo, so the same repo you already have also serves the cart.
func (r *CatalogRepo) ListSellableVariants(ctx context.Context, variantIDs []uuid.UUID) ([]domain.SellableVariant, error) {
	rows, err := r.q.ListSellableVariants(ctx, variantIDs)
	if err != nil {
		return nil, mapCatalogErr(err)
	}

	out := make([]domain.SellableVariant, len(rows))
	for i, row := range rows {
		attrs := map[string]string{}
		if len(row.Attributes) > 0 {
			if err := json.Unmarshal(row.Attributes, &attrs); err != nil {
				return nil, fmt.Errorf("catalog repo: decode attributes of variant %s: %w", row.VariantID, err)
			}
		}
		out[i] = domain.SellableVariant{
			VariantID:      row.VariantID,
			ProductID:      row.ProductID,
			ProductName:    row.ProductName,
			ProductSlug:    row.ProductSlug,
			VariantName:    row.VariantName,
			SKU:            row.Sku,
			Attributes:     attrs,
			ImageURL:       row.ImageUrl, // if your generated field is called ImageUrl, rename it here
			UnitPrice:      row.Price,
			CompareAtPrice: row.CompareAtPrice,
			Available:      int(row.Available),
		}
	}
	return out, nil
}
