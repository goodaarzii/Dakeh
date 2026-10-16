-- +goose Up
-- The client gets a random token. We store only its SHA-256 hash, never the token itself,
-- so a leaked database cannot be used to log in as anyone.
CREATE TABLE refresh_tokens (
                                id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                user_id     UUID         NOT NULL REFERENCES users (id) ON DELETE CASCADE,
                                token_hash  VARCHAR(64)  NOT NULL UNIQUE,          -- SHA-256 as hex = 64 characters
                                user_agent  VARCHAR(255) NOT NULL DEFAULT '',
                                expires_at  TIMESTAMPTZ  NOT NULL,
                                revoked_at  TIMESTAMPTZ,                           -- NULL = still usable
                                created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- "Revoke all tokens of this user" and the cascade delete need this
-- (Postgres does not index foreign keys automatically).
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;