-- Adding a tag that is already there does nothing (safe to repeat).
-- name: AddTagsToProduct :exec
INSERT INTO product_tags (product_id, tag_id)
SELECT sqlc.arg(product_id)::uuid, unnest(sqlc.arg(tag_ids)::uuid[])
    ON CONFLICT DO NOTHING;

-- name: RemoveTagFromProduct :execrows
DELETE FROM product_tags
WHERE product_id = sqlc.arg(product_id) AND tag_id = sqlc.arg(tag_id);

-- Used to replace all tags of a product: clear, then add the new ones, inside one transaction.
-- name: ClearProductTags :exec
DELETE FROM product_tags WHERE product_id = $1;

-- name: ListTagsByProduct :many
SELECT t.*
FROM tags t
         JOIN product_tags pt ON pt.tag_id = t.id
WHERE pt.product_id = $1
ORDER BY t.name;

-- Tags for many products in ONE query (avoids running one query per product in a list).
-- name: ListTagsByProductIDs :many
SELECT pt.product_id, t.id, t.name, t.slug
FROM product_tags pt
         JOIN tags t ON t.id = pt.tag_id
WHERE pt.product_id = ANY(sqlc.arg(product_ids)::uuid[])
ORDER BY pt.product_id, t.name;