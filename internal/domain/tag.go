package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tag is a flat label that can be attached to many products.
type Tag struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
