-- name: CreateProductImage :one
INSERT INTO product_images (product_id, variant_id, storage_key, url, alt_text, sort_order, is_primary)
VALUES (
           sqlc.arg(product_id), sqlc.arg(variant_id), sqlc.arg(storage_key), sqlc.arg(url),
           sqlc.arg(alt_text), sqlc.arg(sort_order), sqlc.arg(is_primary)
       )
    RETURNING *;

-- name: GetProductImageByID :one
SELECT * FROM product_images WHERE id = $1;

-- The primary image first, then by sort_order.
-- name: ListImagesByProduct :many
SELECT * FROM product_images
WHERE product_id = $1
ORDER BY is_primary DESC, sort_order, created_at;

-- To enforce the maximum number of image per product.
-- name: CountImagesByProduct :one
SELECT COUNT(*)::int FROM product_images WHERE product_id = $1;

-- name: UpdateProductImage :one
UPDATE product_images SET
                          alt_text   = sqlc.arg(alt_text),
                          sort_order = sqlc.arg(sort_order)
WHERE id = sqlc.arg(id)
    RETURNING *;

-- Changing the primary image takes TWO steps, inside one transaction:
-- first clear the old one (the unique index allows only one), then set the new one.
-- name: ClearPrimaryImage :exec
UPDATE product_images SET is_primary = FALSE
WHERE product_id = $1 AND is_primary;

-- name: SetPrimaryImage :execrows
UPDATE product_images SET is_primary = TRUE
WHERE id = sqlc.arg(id) AND product_id = sqlc.arg(product_id);

-- Returns the deleted row, so the usecase knows storage_key (to delete the file) and was it primary.
-- name: DeleteProductImage :one
DELETE FROM product_images WHERE id = $1
    RETURNING *;

-- After deleting the primary image: make the first remaining image primary. Does nothing if one exists.
-- name: PromoteNextImage :exec
UPDATE product_images SET is_primary = TRUE
WHERE id = (
    SELECT pi.id FROM product_images pi
    WHERE pi.product_id = $1
    ORDER BY pi.sort_order, pi.created_at
    LIMIT 1
    )
  AND NOT EXISTS (SELECT 1 FROM product_images pi2 WHERE pi2.product_id = $1 AND pi2.is_primary);