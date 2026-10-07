-- name: CreateTag :one
INSERT INTO tags (name, slug)
VALUES (sqlc.arg(name), sqlc.arg(slug))
    RETURNING *;

-- name: GetTagByID :one
SELECT * FROM tags WHERE id = $1;

-- name: GetTagBySlug :one
SELECT * FROM tags WHERE slug = $1;

-- name: ListTags :many
SELECT * FROM tags ORDER BY name;

-- name: UpdateTag :one
UPDATE tags SET
                name       = sqlc.arg(name),
                slug       = sqlc.arg(slug),
                updated_at = NOW()
WHERE id = sqlc.arg(id)
    RETURNING *;

-- Hard delete. ON DELETE CASCADE also removes the tag from every product.
-- name: DeleteTag :execrows
DELETE FROM tags WHERE id = $1;