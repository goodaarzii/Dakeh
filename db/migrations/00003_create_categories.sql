-- +goose Up
CREATE TABLE categories (
                            id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            parent_id    UUID         REFERENCES categories (id) ON DELETE RESTRICT, -- NULL = top-level category
                            name         VARCHAR(100) NOT NULL,
                            slug         VARCHAR(120) NOT NULL UNIQUE,
                            description  TEXT         NOT NULL DEFAULT '',
                            image_url    VARCHAR(500) NOT NULL DEFAULT '',
                            sort_order   INT          NOT NULL DEFAULT 0,
                            is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
                            created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                            updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
                            CHECK (parent_id IS NULL OR parent_id <> id) -- a category cannot be its own parent
);

-- Finding the children of a category. Postgres does not index foreign keys automatically.
CREATE INDEX idx_categories_parent_id ON categories (parent_id);

-- +goose Down
DROP TABLE IF EXISTS categories;