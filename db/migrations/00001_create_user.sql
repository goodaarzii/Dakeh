-- +goose Up
CREATE TABLE users (
                       id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       email          VARCHAR(255) NOT NULL UNIQUE,
                       phone          VARCHAR(20)  NOT NULL UNIQUE,
                       password_hash  VARCHAR(255) NOT NULL,
                       full_name      VARCHAR(255) NOT NULL,
                       role           VARCHAR(20)  NOT NULL DEFAULT 'customer'
                           CHECK (role IN ('customer', 'admin')),
                       is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
                       last_login_at  TIMESTAMPTZ,
                       created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                       updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS users;