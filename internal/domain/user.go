package domain

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleCustomer Role = "customer"
	RoleAdmin    Role = "admin"
)

// User is the business entity. It knows nothing about sqlc, JSON, or Gin.
type User struct {
	ID           uuid.UUID
	Email        string
	Phone        string
	PasswordHash string
	FullName     string
	Role         Role
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
