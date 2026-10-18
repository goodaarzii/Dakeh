-- The LIVE data of cart variants: price, name, image, stock. Redis stores only variant IDs and quantities,
-- so a price changed in the admin is the price in the cart immediately, and a client can never fake a price.
-- Only SELLABLE variants come back: an inactive or deleted variant, or one of a hidden product, is missing from
-- the result, and the usecase treats a missing variant as "no longer available".
-- name: ListSellableVariants :many
SELECT
pv.id                AS variant_id,
pv.sku,
pv.name              AS variant_name,
pv.price,
pv.compare_at_price,
pv.attributes,
p.id                 AS product_id,
p.name               AS product_name,
p.slug               AS product_slug,
COALESCE(i.quantity_on_hand - i.quantity_reserved, 0)::int AS available,
COALESCE((
SELECT pi.url FROM product_images pi
WHERE pi.product_id = p.id AND pi.is_primary
), '')::text         AS image_url
FROM product_variants pv
JOIN products p ON p.id = pv.product_id
LEFT JOIN inventory i ON i.variant_id = pv.id
WHERE pv.id = ANY(sqlc.arg(variant_ids)::uuid[])
AND pv.is_active AND pv.deleted_at IS NULL
AND p.status = 'active' AND p.deleted_at IS NULL;