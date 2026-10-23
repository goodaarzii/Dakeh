-- Favorites are private to the authenticated user. Product visibility is checked when
-- adding a new favorite; the listing query follows the same visibility rules as the shop.

-- name: AddFavorite :execrows
INSERT INTO favorites (user_id, product_id)
SELECT sqlc.arg(user_id), p.id
FROM products p
WHERE p.id = sqlc.arg(product_id)
  AND p.status = 'active'
  AND p.deleted_at IS NULL
  AND EXISTS (
      SELECT 1
      FROM product_variants pv
      WHERE pv.product_id = p.id
        AND pv.is_active
        AND pv.deleted_at IS NULL
  )
ON CONFLICT (user_id, product_id) DO NOTHING;

-- name: IsFavorite :one
SELECT EXISTS (
    SELECT 1
    FROM favorites
    WHERE user_id = $1 AND product_id = $2
) AS is_favorite;

-- name: GetFavoriteCreatedAt :one
SELECT created_at
FROM favorites
WHERE user_id = $1 AND product_id = $2;

-- name: RemoveFavorite :execrows
DELETE FROM favorites
WHERE user_id = $1 AND product_id = $2;

-- One row per visible, sellable product. The returned card data matches GET /products.
-- Favorite records remain saved if a product is temporarily hidden/unavailable; such
-- products are omitted from this list until they become visible in the public catalog.
-- name: ListFavoriteProducts :many
SELECT
    f.product_id,
    f.created_at AS favorited_at,
    p.category_id,
    p.name,
    p.slug,
    p.brand,
    p.avg_rating,
    p.rating_count,
    COALESCE((
        SELECT MIN(pv.price)
        FROM product_variants pv
        WHERE pv.product_id = p.id
          AND pv.is_active
          AND pv.deleted_at IS NULL
    ), 0)::bigint AS min_price,
    EXISTS (
        SELECT 1
        FROM product_variants pv
        JOIN inventory i ON i.variant_id = pv.id
        WHERE pv.product_id = p.id
          AND pv.is_active
          AND pv.deleted_at IS NULL
          AND i.quantity_on_hand - i.quantity_reserved > 0
    ) AS in_stock,
    COALESCE((
        SELECT pi.url
        FROM product_images pi
        WHERE pi.product_id = p.id AND pi.is_primary
    ), '')::text AS image_url
FROM favorites f
JOIN products p ON p.id = f.product_id
WHERE f.user_id = $1
  AND p.status = 'active'
  AND p.deleted_at IS NULL
  AND EXISTS (
      SELECT 1
      FROM product_variants pv
      WHERE pv.product_id = p.id
        AND pv.is_active
        AND pv.deleted_at IS NULL
  )
ORDER BY f.created_at DESC, p.id DESC;
