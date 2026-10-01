package security

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// JWTIssuer creates and verifies access tokens.
// The secret and TTL are injected, so this file never reads env vars.
type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTIssuer(secret string, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret), ttl: ttl}
}

// Issue creates a signed token. The full user UUID goes in the standard "sub" claim.
func (j *JWTIssuer) Issue(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
}

// Parse verifies the signature, the algorithm, and the expiry, then returns the user ID.
// The auth middleware will call this.
func (j *JWTIssuer) Parse(tokenString string) (uuid.UUID, error) {
	var claims jwt.RegisteredClaims
	_, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(*jwt.Token) (any, error) { return j.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.Parse(claims.Subject)
}
