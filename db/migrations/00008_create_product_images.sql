-- +goose Up
-- The file itself lives in object storage (MinIO / S3). The database keeps only where it is.
CREATE TABLE product_images (
                                id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                product_id   UUID         NOT NULL REFERENCES products (id) ON DELETE CASCADE,
                                variant_id   UUID         REFERENCES product_variants (id) ON DELETE SET NULL, -- optional: a photo of one variant
                                storage_key  VARCHAR(255) NOT NULL UNIQUE,   -- the object's name in the bucket, used to delete it
                                url          VARCHAR(500) NOT NULL,          -- the public address, ready to send to the client
                                alt_text     VARCHAR(255) NOT NULL DEFAULT '',
                                sort_order   INT          NOT NULL DEFAULT 0,
                                is_primary   BOOLEAN      NOT NULL DEFAULT FALSE,
                                created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_product_images_product_id ON product_images (product_id);

-- At most ONE primary image per product (the database enforces it).
CREATE UNIQUE INDEX uq_product_images_primary ON product_images (product_id) WHERE is_primary;

-- +goose Down
DROP TABLE IF EXISTS product_images;