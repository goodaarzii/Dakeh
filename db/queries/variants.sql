-- name: CreateVariant :one
INSERT INTO product_variants (
product_id, sku, name, price, compare_at_price, attributes, weight_grams, is_active
) VALUES (
sqlc.arg(product_id), sqlc.arg(sku), sqlc.arg(name), sqlc.arg(price),
sqlc.arg(compare_at_price), sqlc.arg(attributes), sqlc.arg(weight_grams), sqlc.arg(is_active)
)
RETURNING *;

-- name: GetVariantByID :one
SELECT * FROM product_variants
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListVariantsByProduct :many
SELECT * FROM product_variants
WHERE product_id = $1 AND deleted_at IS NULL
ORDER BY created_at, id;

-- Variants of many products in ONE query (for product lists, avoids one query per product).
-- name: ListVariantsByProductIDs :many
SELECT * FROM product_variants
WHERE product_id = ANY(sqlc.arg(product_ids)::uuid[]) AND deleted_at IS NULL
ORDER BY product_id, created_at, id;

-- product_id is not updatable: a variant stays with its product.
-- name: UpdateVariant :one
UPDATE product_variants SET
sku              = sqlc.arg(sku),
name             = sqlc.arg(name),
price            = sqlc.arg(price),
compare_at_price = sqlc.arg(compare_at_price),
attributes       = sqlc.arg(attributes),
weight_grams     = sqlc.arg(weight_grams),
is_active        = sqlc.arg(is_active),
updated_at       = NOW()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteVariant :execrows
UPDATE product_variants
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;