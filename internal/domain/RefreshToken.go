package domain

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken is the record of one issued refresh token.
// The token itself (and its hash) never enters the domain: the hash stays in the repository.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (t RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

func (t RefreshToken) IsExpired(now time.Time) bool {
	return !now.Before(t.ExpiresAt)
}
