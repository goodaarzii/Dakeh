-- name: CreateCategory :one
INSERT INTO categories (parent_id, name, slug, description, image_url, sort_order, is_active)
VALUES (
           sqlc.arg(parent_id), sqlc.arg(name), sqlc.arg(slug), sqlc.arg(description),
           sqlc.arg(image_url), sqlc.arg(sort_order), sqlc.arg(is_active)
       )
    RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = $1;

-- Returns every category. The usecase builds the tree from this flat list.
-- name: ListCategories :many
SELECT * FROM categories ORDER BY sort_order, name;

-- name: UpdateCategory :one
UPDATE categories SET
                      parent_id   = sqlc.arg(parent_id),
                      name        = sqlc.arg(name),
                      slug        = sqlc.arg(slug),
                      description = sqlc.arg(description),
                      image_url   = sqlc.arg(image_url),
                      sort_order  = sqlc.arg(sort_order),
                      is_active   = sqlc.arg(is_active),
                      updated_at  = NOW()
WHERE id = sqlc.arg(id)
    RETURNING *;

-- Fails with a foreign key error (23503) if the category still has products or children.
-- name: DeleteCategory :execrows
DELETE FROM categories WHERE id = $1;