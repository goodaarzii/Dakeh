-- +goose Up
-- A favorite belongs to exactly one user and one product. The composite primary key
-- makes repeated PUTs safe and prevents duplicates at the database level.
CREATE TABLE favorites (
    user_id    UUID        NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    product_id UUID        NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, product_id)
);

-- Listing a user's favorites newest-first should not need a full table scan.
CREATE INDEX idx_favorites_user_created_at
    ON favorites (user_id, created_at DESC, product_id DESC);

-- +goose Down
DROP TABLE IF EXISTS favorites;
