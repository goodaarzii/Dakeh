package domain

import (
	"time"

	"github.com/google/uuid"
)

// Address is a saved delivery address that belongs to one user.
// Soft-deleted addresses never reach the domain, so there is no DeletedAt here.
type Address struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	Label         string // "home", "work"
	RecipientName string
	Phone         string
	Country       string
	Province      string
	City          string
	Street        string
	PostalCode    string
	Lat           float64
	Lng           float64
	IsDefault     bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
