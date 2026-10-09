package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VariantRepo struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewVariantRepo(pool *pgxpool.Pool) *VariantRepo {
	return &VariantRepo{pool: pool, q: sqlc.New(pool)}
}

func (r *VariantRepo) Create(ctx context.Context, variant domain.Variant, initialStock int, lowStockThreshold int) (domain.VariantWithStock, error) {
	attrs, err := attributesToJSON(variant.Attributes)
	if err != nil {
		return domain.VariantWithStock{}, mapVariantErr(err)
	}

	var result domain.VariantWithStock
	err = withTx(ctx, r.pool, func(q *sqlc.Queries) error {
		row, err := q.CreateVariant(ctx, sqlc.CreateVariantParams{
			ProductID:      variant.ProductID,
			Sku:            variant.SKU,
			Name:           variant.Name,
			Price:          variant.Price,
			CompareAtPrice: variant.CompareAtPrice,
			Attributes:     attrs,
			WeightGrams:    int32(variant.WeightGrams),
			IsActive:       variant.IsActive,
		})
		if err != nil {
			return err
		}
		inv, err := q.CreateInventory(ctx, sqlc.CreateInventoryParams{
			VariantID:         row.ID,
			QuantityOnHand:    int32(initialStock),
			LowStockThreshold: int32(lowStockThreshold),
		})
		if err != nil {
			return err
		}
		variant, err := toDomainVariant(row)
		if err != nil {
			return err
		}
		result = domain.VariantWithStock{Variant: variant, Stock: toDomainInventory(inv)}
		return nil
	})
	if err != nil {
		return domain.VariantWithStock{}, mapVariantErr(err)
	}
	return result, nil

}

func (r *VariantRepo) Update(ctx context.Context, v domain.Variant) (domain.Variant, error) {
	attrs, err := attributesToJSON(v.Attributes)
	if err != nil {
		return domain.Variant{}, mapVariantErr(err)
	}

	row, err := r.q.UpdateVariant(ctx, sqlc.UpdateVariantParams{
		ID:             v.ID,
		Sku:            v.SKU,
		Name:           v.Name,
		Price:          v.Price,
		CompareAtPrice: v.CompareAtPrice,
		Attributes:     attrs,
		WeightGrams:    int32(v.WeightGrams),
		IsActive:       v.IsActive,
	})
	if err != nil {
		return domain.Variant{}, mapVariantErr(err) // always return the error, never (empty, nil)
	}
	return toDomainVariant(row)
}

func (r *VariantRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Variant, error) {
	row, err := r.q.GetVariantByID(ctx, id)
	if err != nil {
		return domain.Variant{}, mapVariantErr(err)
	}
	return toDomainVariant(row)
}

func (r *VariantRepo) ListByProduct(ctx context.Context, productID uuid.UUID) ([]domain.Variant, error) {
	rows, err := r.q.ListVariantsByProduct(ctx, productID)
	if err != nil {
		return nil, mapVariantErr(err)
	}
	return toDomainVariants(rows)
}
func (r *VariantRepo) Delete(ctx context.Context, id uuid.UUID) error {
	n, err := r.q.SoftDeleteVariant(ctx, id)
	if err != nil {
		return mapVariantErr(err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *VariantRepo) GetStock(ctx context.Context, variantID uuid.UUID) (domain.Inventory, error) {
	row, err := r.q.GetInventory(ctx, variantID)
	if err != nil {
		return domain.Inventory{}, mapVariantErr(err)
	}
	return toDomainInventory(row), nil
}

// ListStock fetches the stock of many variants in ONE query.
// A variant without a row is simply missing from the result (the usecase handles that).
func (r *VariantRepo) ListStock(ctx context.Context, variantIDs []uuid.UUID) ([]domain.Inventory, error) {
	rows, err := r.q.ListInventoryByVariantIDs(ctx, variantIDs)
	if err != nil {
		return nil, mapVariantErr(err)
	}
	out := make([]domain.Inventory, len(rows))
	for i, row := range rows {
		out[i] = toDomainInventory(row)
	}
	return out, nil
}

// SetStock sets the absolute quantity. A value below "reserved" fails with a CHECK error -> ErrInvalidInput.
func (r *VariantRepo) SetStock(ctx context.Context, variantID uuid.UUID, onHand, lowStockThreshold int) (domain.Inventory, error) {
	row, err := r.q.SetInventory(ctx, sqlc.SetInventoryParams{
		VariantID:         variantID,
		QuantityOnHand:    int32(onHand),
		LowStockThreshold: int32(lowStockThreshold),
	})
	if err != nil {
		return domain.Inventory{}, mapVariantErr(err)
	}
	return toDomainInventory(row), nil
}

// AdjustStock adds (delta > 0) or removes (delta < 0) stock.
// A result below zero or below "reserved" fails with a CHECK error -> ErrInvalidInput.
func (r *VariantRepo) AdjustStock(ctx context.Context, variantID uuid.UUID, delta int) (domain.Inventory, error) {
	row, err := r.q.AdjustInventory(ctx, sqlc.AdjustInventoryParams{
		VariantID: variantID,
		Delta:     int32(delta),
	})
	if err != nil {
		return domain.Inventory{}, mapVariantErr(err)
	}
	return toDomainInventory(row), nil
}

func attributesToJSON(attrs map[string]string) ([]byte, error) {
	if attrs == nil {
		attrs = map[string]string{}
	}
	b, err := json.Marshal(attrs)
	if err != nil {
		return nil, fmt.Errorf("encode attributes: %w", err)
	}
	return b, nil
}

func toDomainVariant(v sqlc.ProductVariant) (domain.Variant, error) {
	attrs := map[string]string{}
	if len(v.Attributes) > 0 {
		if err := json.Unmarshal(v.Attributes, &attrs); err != nil {
			// Fails if someone stored a non-string value (like {"size": 42}) by hand.
			return domain.Variant{}, fmt.Errorf("decode attributes of variant %s: %w", v.ID, err)
		}
	}

	return domain.Variant{
		ID:             v.ID,
		ProductID:      v.ProductID,
		SKU:            v.Sku,
		Name:           v.Name,
		Price:          v.Price,
		CompareAtPrice: v.CompareAtPrice,
		Attributes:     attrs,
		WeightGrams:    int(v.WeightGrams),
		IsActive:       v.IsActive,
		CreatedAt:      v.CreatedAt,
		UpdatedAt:      v.UpdatedAt,
	}, nil
}

func toDomainVariants(rows []sqlc.ProductVariant) ([]domain.Variant, error) {
	out := make([]domain.Variant, len(rows))
	for i, row := range rows {
		v, err := toDomainVariant(row)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

func toDomainInventory(i sqlc.Inventory) domain.Inventory {
	return domain.Inventory{
		VariantID:         i.VariantID,
		QuantityOnHand:    int(i.QuantityOnHand),
		QuantityReserved:  int(i.QuantityReserved),
		LowStockThreshold: int(i.LowStockThreshold),
		UpdatedAt:         i.UpdatedAt,
	}
}

// ---------- error mapping ----------

// mapVariantErr turns database errors into domain errors.
// Add ErrSKUTaken (and ErrInvalidInput if you do not have it yet) to domain/errors.go.
func mapVariantErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique violation
			if pgErr.ConstraintName == "uq_variants_sku" {
				return domain.ErrSKUTaken
			}
		case "23503": // foreign key violation
			if pgErr.ConstraintName == "product_variants_product_id_fkey" {
				return domain.ErrNotFound // the product does not exist
			}
		case "23514": // CHECK violation (negative stock, compare price too low, ...)
			return domain.ErrInvalidInput
		}
	}
	return fmt.Errorf("variant repo: %w", err)
}
