-- +goose Up
CREATE TABLE addresses (
                           id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           user_id         UUID             NOT NULL REFERENCES users (id) ON DELETE CASCADE,
                           label           VARCHAR(50)      NOT NULL DEFAULT '',
                           recipient_name  VARCHAR(255)     NOT NULL,
                           phone           VARCHAR(20)      NOT NULL,
                           country         VARCHAR(100)     NOT NULL,
                           province        VARCHAR(100)     NOT NULL,
                           city            VARCHAR(100)     NOT NULL,
                           street          VARCHAR(500)     NOT NULL,
                           postal_code     VARCHAR(20)      NOT NULL DEFAULT '',
                           lat             DOUBLE PRECISION NOT NULL CHECK (lat BETWEEN -90 AND 90),
                           lng             DOUBLE PRECISION NOT NULL CHECK (lng BETWEEN -180 AND 180),
                           is_default      BOOLEAN          NOT NULL DEFAULT FALSE,
                           created_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
                           updated_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
                           deleted_at      TIMESTAMPTZ
);

-- Listing "my addresses" filters by user_id. Postgres does not index foreign keys automatically.
CREATE INDEX idx_addresses_user_id ON addresses (user_id) WHERE deleted_at IS NULL;

-- A user can have at most one default address (the database enforces it, not only the code).
CREATE UNIQUE INDEX uq_addresses_one_default_per_user
    ON addresses (user_id) WHERE is_default AND deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS addresses;