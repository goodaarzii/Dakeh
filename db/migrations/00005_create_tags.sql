-- +goose Up
CREATE TABLE tags (
                      id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                      name        VARCHAR(50)  NOT NULL UNIQUE,
                      slug        VARCHAR(60)  NOT NULL UNIQUE,
                      created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                      updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Join table: one row = "this product has this tag". It holds the two lines and nothing else.
CREATE TABLE product_tags (
                              product_id  UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
                              tag_id      UUID NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
                              PRIMARY KEY (product_id, tag_id)
);

-- The primary key already makes "tags of a product" fast.
-- This index makes "products of a tag" fast.
CREATE INDEX idx_product_tags_tag_id ON product_tags (tag_id);

-- +goose Down
DROP TABLE IF EXISTS product_tags;
DROP TABLE IF EXISTS tags;