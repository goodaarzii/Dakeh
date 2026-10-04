-- name: CreateProduct :one
INSERT INTO products (category_id, name, slug, description, brand, status)
VALUES (
           sqlc.arg(category_id), sqlc.arg(name), sqlc.arg(slug),
           sqlc.arg(description), sqlc.arg(brand), sqlc.arg(status)
       )
    RETURNING *;

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListProducts :many
SELECT * FROM products
WHERE deleted_at IS NULL
ORDER BY created_at DESC, id DESC
    LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);

-- name: ListProductsByCategory :many
SELECT * FROM products
WHERE category_id = sqlc.arg(category_id) AND deleted_at IS NULL
ORDER BY created_at DESC, id DESC
    LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);

-- name: UpdateProduct :one
UPDATE products SET
                    category_id = sqlc.arg(category_id),
                    name        = sqlc.arg(name),
                    slug        = sqlc.arg(slug),
                    description = sqlc.arg(description),
                    brand       = sqlc.arg(brand),
                    status      = sqlc.arg(status),
                    updated_at  = NOW()
WHERE id = sqlc.arg(id) AND deleted_at IS NULL
    RETURNING *;

-- name: SoftDeleteProduct :execrows
UPDATE products
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;