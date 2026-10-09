-- Called right after CreateVariant, inside the same transaction.
-- name: CreateInventory :one
INSERT INTO inventory (variant_id, quantity_on_hand, low_stock_threshold)
VALUES (sqlc.arg(variant_id), sqlc.arg(quantity_on_hand), sqlc.arg(low_stock_threshold))
    RETURNING *;

-- name: GetInventory :one
SELECT * FROM inventory WHERE variant_id = $1;

-- Stock for many variants in ONE query.
-- name: ListInventoryByVariantIDs :many
SELECT * FROM inventory
WHERE variant_id = ANY(sqlc.arg(variant_ids)::uuid[]);

-- Set the absolute quantity. Fails with a CHECK error (23514) if it would go below reserved.
-- name: SetInventory :one
UPDATE inventory SET
                     quantity_on_hand    = sqlc.arg(quantity_on_hand),
                     low_stock_threshold = sqlc.arg(low_stock_threshold),
                     updated_at          = NOW()
WHERE variant_id = sqlc.arg(variant_id)
    RETURNING *;

-- Add or remove stock: delta +50 = restock, delta -3 = write-off.
-- Fails with a CHECK error (23514) if the result would be negative or below reserved.
-- name: AdjustInventory :one
UPDATE inventory SET
                     quantity_on_hand = quantity_on_hand + sqlc.arg(delta),
                     updated_at       = NOW()
WHERE variant_id = sqlc.arg(variant_id)
    RETURNING *;