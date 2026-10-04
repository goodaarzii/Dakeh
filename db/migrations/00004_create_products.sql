-- +goose Up
CREATE TABLE products (
                          id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          category_id   UUID             NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
                          name          VARCHAR(255)     NOT NULL,
                          slug          VARCHAR(280)     NOT NULL,
                          description   TEXT             NOT NULL DEFAULT '',
                          brand         VARCHAR(100)     NOT NULL DEFAULT '',
                          status        VARCHAR(20)      NOT NULL DEFAULT 'draft'
                              CHECK (status IN ('draft', 'active', 'archived')),
                          avg_rating    DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (avg_rating BETWEEN 0 AND 5),
                          rating_count  INT              NOT NULL DEFAULT 0 CHECK (rating_count >= 0),
                          created_at    TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
                          updated_at    TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
                          deleted_at    TIMESTAMPTZ
);

-- The slug must be unique only among products that are not deleted,
-- so a deleted product does not block reusing its slug.
CREATE UNIQUE INDEX uq_products_slug ON products (slug) WHERE deleted_at IS NULL;

-- Listing products of a category.
CREATE INDEX idx_products_category_id ON products (category_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS products;