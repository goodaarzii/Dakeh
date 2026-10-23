package domain

import (
	"time"

	"github.com/google/uuid"
)

// Favorite records that a user has saved a product for later. The database enforces
// uniqueness for (UserID, ProductID), so saving the same product twice is harmless.
type Favorite struct {
	UserID    uuid.UUID
	ProductID uuid.UUID
	CreatedAt time.Time
}

// FavoriteProduct is a favorite together with the live product-card data needed by the shop.
// Price, image, stock and rating are read from the catalog rather than snapshotted here.
type FavoriteProduct struct {
	Product   ProductCard
	CreatedAt time.Time
}

// NewFavorite creates a valid favorite identity. Product visibility/existence is checked by
// the repository because it depends on current database state.
func NewFavorite(userID, productID uuid.UUID) (Favorite, error) {
	if userID == uuid.Nil || productID == uuid.Nil {
		return Favorite{}, ErrInvalidInput
	}
	return Favorite{UserID: userID, ProductID: productID}, nil
}

// BelongsTo answers whether this favorite belongs to the specified user.
func (f Favorite) BelongsTo(userID uuid.UUID) bool {
	return userID != uuid.Nil && f.UserID == userID
}
