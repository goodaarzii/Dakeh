-- +goose Up
-- Orders preserve a snapshot of the delivery address and product/variant details.
-- Payment is intentionally not integrated yet: an order starts as "submitted".
CREATE TABLE orders (
    id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                  UUID NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    address_id               UUID REFERENCES addresses (id) ON DELETE SET NULL,
    status                   VARCHAR(20) NOT NULL DEFAULT 'submitted'
        CHECK (status IN ('submitted', 'cancelled')),
    idempotency_key          VARCHAR(128) NOT NULL,
    currency                 VARCHAR(3) NOT NULL DEFAULT 'USD',
    subtotal                 BIGINT NOT NULL CHECK (subtotal >= 0),
    shipping_fee             BIGINT NOT NULL CHECK (shipping_fee >= 0),
    total                    BIGINT NOT NULL CHECK (total = subtotal + shipping_fee),
    shipping_label           VARCHAR(50) NOT NULL DEFAULT '',
    shipping_recipient_name  VARCHAR(255) NOT NULL,
    shipping_phone           VARCHAR(20) NOT NULL,
    shipping_country         VARCHAR(100) NOT NULL,
    shipping_province        VARCHAR(100) NOT NULL,
    shipping_city            VARCHAR(100) NOT NULL,
    shipping_street          VARCHAR(500) NOT NULL,
    shipping_postal_code     VARCHAR(20) NOT NULL DEFAULT '',
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    cancelled_at             TIMESTAMPTZ,
    CONSTRAINT uq_orders_user_idempotency_key UNIQUE (user_id, idempotency_key)
);

CREATE INDEX idx_orders_user_created_at
    ON orders (user_id, created_at DESC, id DESC);

CREATE TABLE order_items (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id       UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    -- These IDs are snapshots/reference values, deliberately not foreign keys: order history
    -- must survive if a product or variant is later removed from the catalogue.
    product_id     UUID NOT NULL,
    variant_id     UUID NOT NULL,
    product_name   VARCHAR(255) NOT NULL,
    product_slug   VARCHAR(280) NOT NULL,
    variant_name   VARCHAR(255) NOT NULL DEFAULT '',
    sku            VARCHAR(64) NOT NULL,
    attributes     JSONB NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(attributes) = 'object'),
    image_url      TEXT NOT NULL DEFAULT '',
    unit_price     BIGINT NOT NULL CHECK (unit_price >= 0),
    quantity       INT NOT NULL CHECK (quantity BETWEEN 1 AND 20),
    line_total     BIGINT NOT NULL CHECK (line_total = unit_price * quantity),
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_items_order_id ON order_items (order_id, created_at, id);

CREATE TABLE order_status_history (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id    UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    from_status VARCHAR(20),
    to_status   VARCHAR(20) NOT NULL CHECK (to_status IN ('submitted', 'cancelled')),
    note        VARCHAR(255) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_order_status_history_order_created
    ON order_status_history (order_id, created_at, id);

-- Stock activity is an audit snapshot; variant IDs are intentionally not foreign keys so history survives
-- later catalogue deletion. This phase logs the reservation/release caused by submitted/cancelled orders.
CREATE TABLE stock_movements (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    variant_id     UUID NOT NULL,
    delta_on_hand  INT NOT NULL DEFAULT 0,
    delta_reserved INT NOT NULL DEFAULT 0,
    reason         VARCHAR(40) NOT NULL,
    reference_type VARCHAR(40) NOT NULL DEFAULT 'order',
    reference_id   UUID NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (delta_on_hand <> 0 OR delta_reserved <> 0)
);
CREATE INDEX idx_stock_movements_variant_created
    ON stock_movements (variant_id, created_at DESC, id DESC);
CREATE INDEX idx_stock_movements_reference
    ON stock_movements (reference_type, reference_id);

-- +goose Down
DROP TABLE IF EXISTS stock_movements;
DROP TABLE IF EXISTS order_status_history;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
