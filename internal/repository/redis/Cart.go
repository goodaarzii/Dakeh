package redisrepo

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// CartRepo keeps one Redis HASH per user:
//
//	key   cart:<user id>
//	field <variant id>   value <quantity>
//
// A hash is like a Map<String, Int> stored under one key.
type CartRepo struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCartRepo(client *redis.Client, ttl time.Duration) *CartRepo {
	return &CartRepo{client: client, ttl: ttl}
}

func cartKey(userID uuid.UUID) string {
	return "cart:" + userID.String()
}

// Get returns every item of the cart. A cart that does not exist is an empty list, not an error.
func (r *CartRepo) Get(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error) {
	m, err := r.client.HGetAll(ctx, cartKey(userID)).Result()
	if err != nil {
		return nil, fmt.Errorf("cart repo: get: %w", err)
	}

	items := make([]domain.CartItem, 0, len(m))
	for field, value := range m {
		id, errID := uuid.Parse(field)
		qty, errQty := strconv.Atoi(value)
		if errID != nil || errQty != nil || qty < 1 {
			continue // skip garbage: bad data in Redis must never crash a request
		}
		items = append(items, domain.CartItem{VariantID: id, Quantity: qty})
	}
	return items, nil
}

// SetItems adds or replaces items and renews the expiry: a cart lives CartTTL after its LAST change.
// HSET and EXPIRE are sent together in one transaction (a pipeline).
func (r *CartRepo) SetItems(ctx context.Context, userID uuid.UUID, items []domain.CartItem) error {
	if len(items) == 0 {
		return nil
	}
	key := cartKey(userID)

	pipe := r.client.TxPipeline()
	for _, it := range items {
		pipe.HSet(ctx, key, it.VariantID.String(), it.Quantity)
	}
	pipe.Expire(ctx, key, r.ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("cart repo: set items: %w", err)
	}
	return nil
}

// RemoveItem deletes one line. A line that is not there is not an error (idempotent).
func (r *CartRepo) RemoveItem(ctx context.Context, userID, variantID uuid.UUID) error {
	if err := r.client.HDel(ctx, cartKey(userID), variantID.String()).Err(); err != nil {
		return fmt.Errorf("cart repo: remove item: %w", err)
	}
	return nil
}

// Clear deletes the whole cart.
func (r *CartRepo) Clear(ctx context.Context, userID uuid.UUID) error {
	if err := r.client.Del(ctx, cartKey(userID)).Err(); err != nil {
		return fmt.Errorf("cart repo: clear: %w", err)
	}
	return nil
}
