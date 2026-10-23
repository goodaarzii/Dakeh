-- Resolve only products that are currently visible in the public catalog.
-- name: GetPublicProductIDByID :one
SELECT p.id
FROM products p
WHERE p.id = sqlc.arg(product_id)
  AND p.status = 'active'
  AND p.deleted_at IS NULL
  AND EXISTS (
      SELECT 1 FROM product_variants pv
      WHERE pv.product_id = p.id
        AND pv.is_active
        AND pv.deleted_at IS NULL
  );

-- name: ListReviews :many
SELECT
    r.*,
    u.full_name AS reviewer_name
FROM reviews r
         JOIN users u ON u.id = r.user_id
WHERE r.user_id = $1
ORDER BY r.created_at DESC
    LIMIT $2 OFFSET $3;

-- Public product reviews, newest first.
-- name: ListReviewsByProduct :many
SELECT
    r.id,
    r.product_id,
    r.user_id,
    u.full_name AS reviewer_name,
    r.rating::int AS rating,
    r.title,
    r.body,
    r.created_at,
    r.updated_at
FROM reviews r
JOIN users u ON u.id = r.user_id
WHERE r.product_id = sqlc.arg(product_id)
ORDER BY r.created_at DESC, r.id DESC
LIMIT sqlc.arg(page_size)::int
OFFSET sqlc.arg(page_offset)::int;

-- name: CountReviewsByProduct :one
SELECT COUNT(*)::bigint
FROM reviews
WHERE product_id = sqlc.arg(product_id);

-- Used by the logged-in reviewer to find whether they have an existing review for this product.
-- name: GetMyReviewForProduct :one
SELECT
    r.id,
    r.product_id,
    r.user_id,
    u.full_name AS reviewer_name,
    r.rating::int AS rating,
    r.title,
    r.body,
    r.created_at,
    r.updated_at
FROM reviews r
JOIN users u ON u.id = r.user_id
WHERE r.product_id = sqlc.arg(product_id)
  AND r.user_id = sqlc.arg(user_id);

-- Create only. uq_reviews_product_user rejects a second review for the same user/product.
-- name: CreateProductReview :one
WITH target_product AS (
    SELECT p.id
    FROM products p
    WHERE p.id = sqlc.arg(product_id)
      AND p.status = 'active'
      AND p.deleted_at IS NULL
      AND EXISTS (
          SELECT 1 FROM product_variants pv
          WHERE pv.product_id = p.id
            AND pv.is_active
            AND pv.deleted_at IS NULL
      )
), created AS (
    INSERT INTO reviews (product_id, user_id, rating, title, body)
    SELECT
        target_product.id,
        sqlc.arg(user_id)::uuid,
        sqlc.arg(rating)::smallint,
        sqlc.arg(title)::varchar,
        sqlc.arg(body)::text
    FROM target_product
    WHERE TRUE
    RETURNING *
)
SELECT
    created.id,
    created.product_id,
    created.user_id,
    u.full_name AS reviewer_name,
    created.rating::int AS rating,
    created.title,
    created.body,
    created.created_at,
    created.updated_at
FROM created
JOIN users u ON u.id = created.user_id;

-- Update/delete always include user_id from the token, so users cannot modify another user's review.
-- name: UpdateMyReview :one
WITH updated AS (
UPDATE reviews AS r
SET rating = sqlc.arg(rating)::smallint,
        title = sqlc.arg(title)::varchar,
        body = sqlc.arg(body)::text,
        updated_at = CASE
            WHEN r.rating IS DISTINCT FROM sqlc.arg(rating)::smallint
    OR r.title IS DISTINCT FROM sqlc.arg(title)::varchar
    OR r.body IS DISTINCT FROM sqlc.arg(body)::text
    THEN NOW()
    ELSE r.updated_at
END
WHERE r.id = sqlc.arg(review_id)
      AND r.user_id = sqlc.arg(user_id)
    RETURNING r.*
)
SELECT
    updated.id,
    updated.product_id,
    updated.user_id,
    u.full_name AS reviewer_name,
    updated.rating::int AS rating,
    updated.title,
    updated.body,
    updated.created_at,
    updated.updated_at
FROM updated
         JOIN users u ON u.id = updated.user_id;

-- name: DeleteMyReview :execrows
DELETE FROM reviews
WHERE id = sqlc.arg(review_id)
  AND user_id = sqlc.arg(user_id);
