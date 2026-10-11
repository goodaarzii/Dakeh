-- Public (customer-facing) reads. Only visible things: active products, active variants, active categories.

-- name: ListActiveCategories :many
SELECT * FROM categories
WHERE is_active
ORDER BY sort_order, name;

-- name: GetActiveProductBySlug :one
SELECT * FROM products
WHERE slug = $1 AND status = 'active' AND deleted_at IS NULL;

-- Sellable variants only. (The admin ListVariantsByProduct does not filter on is_active.)
-- name: ListActiveVariantsByProduct :many
SELECT * FROM product_variants
WHERE product_id = $1 AND is_active AND deleted_at IS NULL
ORDER BY created_at, id;

-- Product cards for the shop pages. ONE query returns what a card needs:
-- the lowest variant price ("from $X") and whether anything is in stock.
-- Empty values mean "no filter": category_ids = '{}', tag_slug = '', search = ''.
-- sort is 'newest' (default), 'price_asc' or 'price_desc'.
-- name: ListPublicProducts :many
WITH visible AS (
    SELECT
        p.id,
        p.category_id,
        p.name,
        p.slug,
        p.brand,
        p.avg_rating,
        p.rating_count,
        p.created_at,
        COALESCE((
                     SELECT MIN(pv.price)
                     FROM product_variants pv
                     WHERE pv.product_id = p.id AND pv.is_active AND pv.deleted_at IS NULL
                 ), 0)::bigint AS min_price,
        EXISTS (
            SELECT 1
            FROM product_variants pv
                     JOIN inventory i ON i.variant_id = pv.id
            WHERE pv.product_id = p.id AND pv.is_active AND pv.deleted_at IS NULL
              AND i.quantity_on_hand - i.quantity_reserved > 0
        ) AS in_stock
    FROM products p
    WHERE p.status = 'active'
      AND p.deleted_at IS NULL
      -- a product with no sellable variant is not shown
      AND EXISTS (
        SELECT 1 FROM product_variants pv
        WHERE pv.product_id = p.id AND pv.is_active AND pv.deleted_at IS NULL
    )
      AND (cardinality(sqlc.arg(category_ids)::uuid[]) = 0
        OR p.category_id = ANY(sqlc.arg(category_ids)::uuid[]))
      AND (sqlc.arg(tag_slug)::text = '' OR EXISTS (
            SELECT 1 FROM product_tags pt
            JOIN tags t ON t.id = pt.tag_id
            WHERE pt.product_id = p.id AND t.slug = sqlc.arg(tag_slug)::text
      ))
      AND (sqlc.arg(search)::text = '' OR p.name ILIKE '%' || sqlc.arg(search)::text || '%')
)
SELECT * FROM visible
ORDER BY
    CASE WHEN sqlc.arg(sort)::text = 'price_asc'  THEN min_price END ASC,
    CASE WHEN sqlc.arg(sort)::text = 'price_desc' THEN min_price END DESC,
    created_at DESC, id DESC
    LIMIT sqlc.arg(page_size) OFFSET sqlc.arg(page_offset);