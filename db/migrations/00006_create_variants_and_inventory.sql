-- +goose Up
-- A product is the "idea" (Running Shoe X). A variant is what you actually sell
-- (Running Shoe X, red, size 42). Price, SKU and stock live on the variant.
CREATE TABLE product_variants (
                                  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                  product_id       UUID         NOT NULL REFERENCES products (id) ON DELETE CASCADE,
                                  sku              VARCHAR(64)  NOT NULL,
                                  name             VARCHAR(255) NOT NULL DEFAULT '',          -- e.g. "Red / 42"
                                  price            BIGINT       NOT NULL CHECK (price >= 0),   -- USD cents: 1999 = $19.99
                                  compare_at_price BIGINT       NOT NULL DEFAULT 0,            -- old price, 0 = none
                                  attributes       JSONB        NOT NULL DEFAULT '{}'::jsonb,  -- {"color":"red","size":"42"}
                                  weight_grams     INT          NOT NULL DEFAULT 0 CHECK (weight_grams >= 0),
                                  is_active        BOOLEAN      NOT NULL DEFAULT TRUE,
                                  created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                                  updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                                  deleted_at       TIMESTAMPTZ,
                                  CHECK (compare_at_price = 0 OR compare_at_price > price),
                                  CHECK (jsonb_typeof(attributes) = 'object')
);

-- SKU must be unique only among variants that are not deleted.
CREATE UNIQUE INDEX uq_variants_sku ON product_variants (sku) WHERE deleted_at IS NULL;

-- Listing the variants of a product.
CREATE INDEX idx_variants_product_id ON product_variants (product_id) WHERE deleted_at IS NULL;

-- Stock: one row per variant (1:1).
-- available = quantity_on_hand - quantity_reserved
CREATE TABLE inventory (
                           variant_id          UUID PRIMARY KEY REFERENCES product_variants (id) ON DELETE CASCADE,
                           quantity_on_hand    INT         NOT NULL DEFAULT 0 CHECK (quantity_on_hand >= 0),
                           quantity_reserved   INT         NOT NULL DEFAULT 0 CHECK (quantity_reserved >= 0),
                           low_stock_threshold INT         NOT NULL DEFAULT 0 CHECK (low_stock_threshold >= 0),
                           updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    -- The database itself refuses to reserve more than we have (no overselling).
                           CHECK (quantity_reserved <= quantity_on_hand)
);

-- +goose Down
DROP TABLE IF EXISTS inventory;
DROP TABLE IF EXISTS product_variants;