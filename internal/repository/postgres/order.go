package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/goodaarzii/dakeh/internal/domain"
)

type OrderRepo struct {
	pool *pgxpool.Pool
}

func NewOrderRepo(pool *pgxpool.Pool) *OrderRepo {
	return &OrderRepo{pool: pool}
}

type checkoutVariant struct {
	VariantID      uuid.UUID
	ProductID      uuid.UUID
	ProductName    string
	ProductSlug    string
	VariantName    string
	SKU            string
	Attributes     map[string]string
	ImageURL       string
	UnitPrice      int64
	CompareAtPrice int64
	Available      int
}

// Create re-reads all sellable product data and reserves stock under row locks.
// The order, snapshots, and inventory reservations commit or roll back together.
func (r *OrderRepo) Create(ctx context.Context, userID, addressID uuid.UUID, idempotencyKey string, cartItems []domain.CartItem, shippingFee int64) (domain.Order, error) {
	if userID == uuid.Nil || addressID == uuid.Nil || idempotencyKey == "" || len(idempotencyKey) > 128 || len(cartItems) == 0 || len(cartItems) > domain.MaxCartLines || shippingFee < 0 {
		return domain.Order{}, domain.ErrInvalidInput
	}

	wanted := make(map[uuid.UUID]int, len(cartItems))
	ids := make([]uuid.UUID, 0, len(cartItems))
	for _, item := range cartItems {
		if item.VariantID == uuid.Nil || item.Quantity < 1 || item.Quantity > domain.MaxQuantityPerLine {
			return domain.Order{}, domain.ErrInvalidInput
		}
		if _, exists := wanted[item.VariantID]; exists {
			return domain.Order{}, domain.ErrInvalidInput
		}
		wanted[item.VariantID] = item.Quantity
		ids = append(ids, item.VariantID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, fmt.Errorf("order repo: begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var address domain.OrderAddress
	err = tx.QueryRow(ctx, `
SELECT label, recipient_name, phone, country, province, city, street, postal_code
FROM addresses
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
FOR SHARE`, addressID, userID).Scan(
		&address.Label, &address.RecipientName, &address.Phone, &address.Country,
		&address.Province, &address.City, &address.Street, &address.PostalCode,
	)
	if err != nil {
		return domain.Order{}, mapOrderErr(err)
	}

	rows, err := tx.Query(ctx, `
SELECT
    pv.id,
    p.id,
    p.name,
    p.slug,
    pv.name,
    pv.sku,
    pv.attributes,
    COALESCE((SELECT pi.url FROM product_images pi WHERE pi.product_id = p.id AND pi.is_primary LIMIT 1), '')::text,
    pv.price,
    pv.compare_at_price,
    (i.quantity_on_hand - i.quantity_reserved)::int
FROM product_variants pv
JOIN products p ON p.id = pv.product_id
JOIN inventory i ON i.variant_id = pv.id
WHERE pv.id = ANY($1::uuid[])
  AND pv.is_active AND pv.deleted_at IS NULL
  AND p.status = 'active' AND p.deleted_at IS NULL
ORDER BY pv.id
FOR UPDATE OF pv, p, i`, ids)
	if err != nil {
		return domain.Order{}, fmt.Errorf("order repo: load sellable variants: %w", err)
	}

	variants := make(map[uuid.UUID]checkoutVariant, len(ids))
	for rows.Next() {
		var v checkoutVariant
		var attrs []byte
		if err := rows.Scan(
			&v.VariantID, &v.ProductID, &v.ProductName, &v.ProductSlug,
			&v.VariantName, &v.SKU, &attrs, &v.ImageURL,
			&v.UnitPrice, &v.CompareAtPrice, &v.Available,
		); err != nil {
			rows.Close()
			return domain.Order{}, fmt.Errorf("order repo: scan checkout variant: %w", err)
		}
		v.Attributes = map[string]string{}
		if len(attrs) > 0 {
			if err := json.Unmarshal(attrs, &v.Attributes); err != nil {
				rows.Close()
				return domain.Order{}, fmt.Errorf("order repo: decode variant attributes: %w", err)
			}
		}
		variants[v.VariantID] = v
	}
	rowsErr := rows.Err()
	rows.Close()
	if rowsErr != nil {
		return domain.Order{}, fmt.Errorf("order repo: iterate checkout variants: %w", rowsErr)
	}
	if len(variants) != len(wanted) {
		return domain.Order{}, domain.ErrVariantUnavailable
	}

	const maxMoney = int64(math.MaxInt64)
	var subtotal int64
	orderItems := make([]domain.OrderItem, 0, len(cartItems))
	for _, variantID := range ids {
		v := variants[variantID]
		quantity := wanted[variantID]
		if quantity > v.Available {
			return domain.Order{}, domain.ErrInsufficientStock
		}
		if v.UnitPrice > 0 && int64(quantity) > maxMoney/v.UnitPrice {
			return domain.Order{}, domain.ErrInvalidInput
		}
		lineTotal := v.UnitPrice * int64(quantity)
		if subtotal > maxMoney-lineTotal {
			return domain.Order{}, domain.ErrInvalidInput
		}
		subtotal += lineTotal
		orderItems = append(orderItems, domain.OrderItem{
			ProductID: v.ProductID, VariantID: v.VariantID,
			ProductName: v.ProductName, ProductSlug: v.ProductSlug,
			VariantName: v.VariantName, SKU: v.SKU, Attributes: v.Attributes,
			ImageURL: v.ImageURL, UnitPrice: v.UnitPrice,
			Quantity: quantity, LineTotal: lineTotal,
		})
	}
	if subtotal > maxMoney-shippingFee {
		return domain.Order{}, domain.ErrInvalidInput
	}
	total := subtotal + shippingFee

	var order domain.Order
	order.UserID = userID
	order.Status = domain.OrderStatusSubmitted
	order.Currency = "USD"
	order.Subtotal = subtotal
	order.ShippingFee = shippingFee
	order.Total = total
	order.Address = address
	order.Items = make([]domain.OrderItem, len(orderItems))
	err = tx.QueryRow(ctx, `
INSERT INTO orders (
    user_id, address_id, status, idempotency_key, currency, subtotal, shipping_fee, total,
    shipping_label, shipping_recipient_name, shipping_phone, shipping_country,
    shipping_province, shipping_city, shipping_street, shipping_postal_code
) VALUES ($1, $2, 'submitted', $3, 'USD', $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
RETURNING id, created_at, updated_at`,
		userID, addressID, idempotencyKey, subtotal, shippingFee, total,
		address.Label, address.RecipientName, address.Phone, address.Country,
		address.Province, address.City, address.Street, address.PostalCode,
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return domain.Order{}, mapOrderErr(err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO order_status_history (order_id, from_status, to_status, note)
VALUES ($1, NULL, 'submitted', 'Order submitted; payment not collected')`, order.ID); err != nil {
		return domain.Order{}, fmt.Errorf("order repo: create initial status history: %w", err)
	}

	for i, item := range orderItems {
		attributes, err := json.Marshal(item.Attributes)
		if err != nil {
			return domain.Order{}, fmt.Errorf("order repo: encode order item attributes: %w", err)
		}
		var orderItemID uuid.UUID
		var createdAt time.Time
		err = tx.QueryRow(ctx, `
INSERT INTO order_items (
    order_id, product_id, variant_id, product_name, product_slug, variant_name,
    sku, attributes, image_url, unit_price, quantity, line_total
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10, $11, $12)
RETURNING id, created_at`,
			order.ID, item.ProductID, item.VariantID, item.ProductName, item.ProductSlug,
			item.VariantName, item.SKU, attributes, item.ImageURL,
			item.UnitPrice, item.Quantity, item.LineTotal,
		).Scan(&orderItemID, &createdAt)
		if err != nil {
			return domain.Order{}, mapOrderErr(err)
		}
		item.ID = orderItemID
		item.OrderID = order.ID
		item.CreatedAt = createdAt
		order.Items[i] = item

		reserved, err := tx.Exec(ctx, `
UPDATE inventory
SET quantity_reserved = quantity_reserved + $2, updated_at = NOW()
WHERE variant_id = $1 AND quantity_on_hand - quantity_reserved >= $2`, item.VariantID, item.Quantity)
		if err != nil {
			return domain.Order{}, fmt.Errorf("order repo: reserve inventory: %w", err)
		}
		if reserved.RowsAffected() != 1 {
			return domain.Order{}, domain.ErrInsufficientStock
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO stock_movements (variant_id, delta_on_hand, delta_reserved, reason, reference_type, reference_id)
VALUES ($1, 0, $2, 'order_submitted', 'order', $3)`, item.VariantID, item.Quantity, order.ID); err != nil {
			return domain.Order{}, fmt.Errorf("order repo: record inventory reservation: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("order repo: commit: %w", err)
	}
	return order, nil
}

func (r *OrderRepo) FindByIdempotencyKey(ctx context.Context, userID uuid.UUID, key string) (domain.Order, bool, error) {
	var orderID uuid.UUID
	err := r.pool.QueryRow(ctx, `SELECT id FROM orders WHERE user_id = $1 AND idempotency_key = $2`, userID, key).Scan(&orderID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Order{}, false, nil
	}
	if err != nil {
		return domain.Order{}, false, fmt.Errorf("order repo: lookup idempotency key: %w", err)
	}
	item, err := r.getOne(ctx, userID, orderID)
	if err != nil {
		return domain.Order{}, false, err
	}
	return item, true, nil
}

func (r *OrderRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Order, int, error) {
	var total int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("order repo: count orders: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
SELECT id, user_id, status, currency, subtotal, shipping_fee, total,
       shipping_label, shipping_recipient_name, shipping_phone, shipping_country,
       shipping_province, shipping_city, shipping_street, shipping_postal_code,
       created_at, updated_at, cancelled_at
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC, id DESC
LIMIT $2 OFFSET $3`, userID, int32(limit), int32(offset))
	if err != nil {
		return nil, int(total), fmt.Errorf("order repo: list orders: %w", err)
	}
	orders := make([]domain.Order, 0, limit)
	ids := make([]uuid.UUID, 0, limit)
	for rows.Next() {
		item, err := scanOrder(rows)
		if err != nil {
			rows.Close()
			return nil, int(total), fmt.Errorf("order repo: scan order: %w", err)
		}
		orders = append(orders, item)
		ids = append(ids, item.ID)
	}
	rowsErr := rows.Err()
	rows.Close()
	if rowsErr != nil {
		return nil, int(total), fmt.Errorf("order repo: iterate orders: %w", rowsErr)
	}
	if len(ids) == 0 {
		return orders, int(total), nil
	}

	itemRows, err := r.pool.Query(ctx, `
SELECT id, order_id, product_id, variant_id, product_name, product_slug,
       variant_name, sku, attributes, image_url, unit_price, quantity, line_total, created_at
FROM order_items
WHERE order_id = ANY($1::uuid[])
ORDER BY order_id, created_at, id`, ids)
	if err != nil {
		return nil, int(total), fmt.Errorf("order repo: list order items: %w", err)
	}
	itemsByOrder := make(map[uuid.UUID][]domain.OrderItem, len(ids))
	for itemRows.Next() {
		item, err := scanOrderItem(itemRows)
		if err != nil {
			itemRows.Close()
			return nil, int(total), fmt.Errorf("order repo: scan order item: %w", err)
		}
		itemsByOrder[item.OrderID] = append(itemsByOrder[item.OrderID], item)
	}
	itemsErr := itemRows.Err()
	itemRows.Close()
	if itemsErr != nil {
		return nil, int(total), fmt.Errorf("order repo: iterate order items: %w", itemsErr)
	}
	for i := range orders {
		orders[i].Items = itemsByOrder[orders[i].ID]
		if orders[i].Items == nil {
			orders[i].Items = []domain.OrderItem{}
		}
	}
	return orders, int(total), nil
}

func (r *OrderRepo) Cancel(ctx context.Context, userID, orderID uuid.UUID) (domain.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return domain.Order{}, fmt.Errorf("order repo: begin cancel transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1 AND user_id = $2 FOR UPDATE`, orderID, userID).Scan(&status)
	if err != nil {
		return domain.Order{}, mapOrderErr(err)
	}
	if status != domain.OrderStatusSubmitted {
		return domain.Order{}, domain.ErrOrderCannotCancel
	}

	rows, err := tx.Query(ctx, `SELECT variant_id, quantity FROM order_items WHERE order_id = $1 ORDER BY variant_id`, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("order repo: load items to release stock: %w", err)
	}
	type reservation struct {
		variantID uuid.UUID
		quantity  int
	}
	reservations := make([]reservation, 0, 8)
	for rows.Next() {
		var item reservation
		if err := rows.Scan(&item.variantID, &item.quantity); err != nil {
			rows.Close()
			return domain.Order{}, fmt.Errorf("order repo: scan reservation: %w", err)
		}
		reservations = append(reservations, item)
	}
	rowsErr := rows.Err()
	rows.Close()
	if rowsErr != nil {
		return domain.Order{}, fmt.Errorf("order repo: iterate reservations: %w", rowsErr)
	}
	for _, item := range reservations {
		if _, err := tx.Exec(ctx, `
UPDATE inventory
SET quantity_reserved = GREATEST(0, quantity_reserved - $2), updated_at = NOW()
WHERE variant_id = $1`, item.variantID, item.quantity); err != nil {
			return domain.Order{}, fmt.Errorf("order repo: release inventory: %w", err)
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO stock_movements (variant_id, delta_on_hand, delta_reserved, reason, reference_type, reference_id)
VALUES ($1, 0, $2, 'order_cancelled', 'order', $3)`, item.variantID, -item.quantity, orderID); err != nil {
			return domain.Order{}, fmt.Errorf("order repo: record inventory release: %w", err)
		}
	}
	if _, err := tx.Exec(ctx, `
UPDATE orders SET status = 'cancelled', cancelled_at = NOW(), updated_at = NOW()
WHERE id = $1 AND user_id = $2`, orderID, userID); err != nil {
		return domain.Order{}, fmt.Errorf("order repo: mark order cancelled: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO order_status_history (order_id, from_status, to_status, note)
VALUES ($1, 'submitted', 'cancelled', 'Customer cancelled before payment')`, orderID); err != nil {
		return domain.Order{}, fmt.Errorf("order repo: record cancellation history: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Order{}, fmt.Errorf("order repo: commit cancellation: %w", err)
	}
	return r.getOne(ctx, userID, orderID)
}

func (r *OrderRepo) getOne(ctx context.Context, userID, orderID uuid.UUID) (domain.Order, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, user_id, status, currency, subtotal, shipping_fee, total,
       shipping_label, shipping_recipient_name, shipping_phone, shipping_country,
       shipping_province, shipping_city, shipping_street, shipping_postal_code,
       created_at, updated_at, cancelled_at
FROM orders WHERE user_id = $1 AND id = $2`, userID, orderID)
	order, err := scanOrderRow(row)
	if err != nil {
		return domain.Order{}, mapOrderErr(err)
	}
	rows, err := r.pool.Query(ctx, `
SELECT id, order_id, product_id, variant_id, product_name, product_slug,
       variant_name, sku, attributes, image_url, unit_price, quantity, line_total, created_at
FROM order_items WHERE order_id = $1 ORDER BY created_at, id`, orderID)
	if err != nil {
		return domain.Order{}, fmt.Errorf("order repo: get order items: %w", err)
	}
	defer rows.Close()
	order.Items = make([]domain.OrderItem, 0, 4)
	for rows.Next() {
		item, err := scanOrderItem(rows)
		if err != nil {
			return domain.Order{}, fmt.Errorf("order repo: scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	if err := rows.Err(); err != nil {
		return domain.Order{}, fmt.Errorf("order repo: iterate order items: %w", err)
	}
	return order, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanOrder(rows pgx.Rows) (domain.Order, error) {
	return scanOrderRow(rows)
}

func scanOrderRow(row rowScanner) (domain.Order, error) {
	var order domain.Order
	var cancelledAt pgtype.Timestamptz
	err := row.Scan(
		&order.ID, &order.UserID, &order.Status, &order.Currency,
		&order.Subtotal, &order.ShippingFee, &order.Total,
		&order.Address.Label, &order.Address.RecipientName, &order.Address.Phone,
		&order.Address.Country, &order.Address.Province, &order.Address.City,
		&order.Address.Street, &order.Address.PostalCode,
		&order.CreatedAt, &order.UpdatedAt, &cancelledAt,
	)
	if err != nil {
		return domain.Order{}, err
	}
	if cancelledAt.Valid {
		value := cancelledAt.Time
		order.CancelledAt = &value
	}
	order.Items = []domain.OrderItem{}
	return order, nil
}

func scanOrderItem(row rowScanner) (domain.OrderItem, error) {
	var item domain.OrderItem
	var attrs []byte
	err := row.Scan(
		&item.ID, &item.OrderID, &item.ProductID, &item.VariantID,
		&item.ProductName, &item.ProductSlug, &item.VariantName, &item.SKU,
		&attrs, &item.ImageURL, &item.UnitPrice, &item.Quantity, &item.LineTotal, &item.CreatedAt,
	)
	if err != nil {
		return domain.OrderItem{}, err
	}
	item.Attributes = map[string]string{}
	if len(attrs) > 0 {
		if err := json.Unmarshal(attrs, &item.Attributes); err != nil {
			return domain.OrderItem{}, fmt.Errorf("decode item attributes: %w", err)
		}
	}
	return item, nil
}

func mapOrderErr(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	code, constraint := pgConstraint(err)
	switch code {
	case "23505":
		if constraint == "uq_orders_user_idempotency_key" {
			return domain.ErrConflict
		}
	case "23503":
		if constraint == "orders_address_id_fkey" || constraint == "orders_user_id_fkey" {
			return domain.ErrNotFound
		}
	case "23514":
		return domain.ErrInvalidInput
	}
	return fmt.Errorf("order repo: %w", err)
}
