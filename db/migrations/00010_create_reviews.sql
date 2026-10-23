-- +goose Up
-- A user may leave exactly one review per product. The review's rating is the
-- source of truth; products.avg_rating and products.rating_count are maintained
-- incrementally by the trigger below so existing catalog responses stay useful.
CREATE TABLE reviews (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    user_id    UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    rating     SMALLINT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title      VARCHAR(120) NOT NULL DEFAULT '',
    body       TEXT NOT NULL CHECK (char_length(btrim(body)) BETWEEN 3 AND 5000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_reviews_product_user UNIQUE (product_id, user_id)
);

CREATE INDEX idx_reviews_product_created_at
    ON reviews (product_id, created_at DESC, id DESC);

-- Existing product rating values had no review source, so reset them to the
-- accurate empty state before review rows start contributing to the aggregate.
UPDATE products SET avg_rating = 0, rating_count = 0;

-- Keep the product's existing rating summary in sync without an aggregate scan
-- on every write. The arithmetic uses the old total/count and is updated atomically
-- on the product row, including when concurrent reviews target the same product.
-- Keep the product's existing rating summary in sync without an aggregate scan
-- on every write.

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION sync_product_review_rating_stats()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE products
        SET avg_rating = ((avg_rating * rating_count) + NEW.rating::double precision)
                         / (rating_count + 1),
            rating_count = rating_count + 1
        WHERE id = NEW.product_id;
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE products
        SET avg_rating = CASE
                WHEN rating_count <= 1 THEN 0
                ELSE ((avg_rating * rating_count) - OLD.rating::double precision)
                     / (rating_count - 1)
            END,
            rating_count = GREATEST(rating_count - 1, 0)
        WHERE id = OLD.product_id;
        RETURN OLD;
    ELSE
        IF OLD.product_id = NEW.product_id THEN
            IF OLD.rating IS DISTINCT FROM NEW.rating THEN
                UPDATE products
                SET avg_rating = CASE
                        WHEN rating_count = 0 THEN 0
                        ELSE ((avg_rating * rating_count)
                              - OLD.rating::double precision
                              + NEW.rating::double precision) / rating_count
                    END
                WHERE id = NEW.product_id;
            END IF;
        ELSE
            UPDATE products
            SET avg_rating = CASE
                    WHEN rating_count <= 1 THEN 0
                    ELSE ((avg_rating * rating_count) - OLD.rating::double precision)
                         / (rating_count - 1)
                END,
                rating_count = GREATEST(rating_count - 1, 0)
            WHERE id = OLD.product_id;

            UPDATE products
            SET avg_rating = ((avg_rating * rating_count) + NEW.rating::double precision)
                             / (rating_count + 1),
                rating_count = rating_count + 1
            WHERE id = NEW.product_id;
        END IF;
        RETURN NEW;
END IF;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER reviews_sync_product_rating_stats
AFTER INSERT OR UPDATE OR DELETE ON reviews
FOR EACH ROW EXECUTE FUNCTION sync_product_review_rating_stats();

-- +goose Down
DROP TRIGGER IF EXISTS reviews_sync_product_rating_stats ON reviews;
DROP FUNCTION IF EXISTS sync_product_review_rating_stats();
DROP TABLE IF EXISTS reviews;
-- Keep avg_rating/rating_count at 0 after rollback because review records no longer exist.
UPDATE products SET avg_rating = 0, rating_count = 0;
