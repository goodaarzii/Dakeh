package domain

import "errors"

// Business errors. Repositories translate DB errors into these,
// and the transport layer translates these into HTTP status codes.
var (
	ErrNotFound            = errors.New("not found")
	ErrEmailTaken          = errors.New("email already registered")
	ErrPhoneTaken          = errors.New("phone already registered")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUserInactive        = errors.New("user is inactive")
	ErrSKUTaken            = errors.New("sku is taken")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInvalidRefreshToken = errors.New("user token expired")
)
