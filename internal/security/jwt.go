package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// Claims is our own claims struct. It embeds the standard claims (sub, exp, iat)
// and adds the custom "role" field. This is how you add extra data to a token.
type Claims struct {
	Role domain.Role `json:"role"`
	jwt.RegisteredClaims
}

// JWTIssuer creates and verifies access tokens.
// The secret and TTL are injected, so this file never reads env vars.
type JWTIssuer struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTIssuer(secret string, ttl time.Duration) *JWTIssuer {
	return &JWTIssuer{secret: []byte(secret), ttl: ttl}
}

// Issue creates a signed token: user ID in "sub", role in "role".
func (j *JWTIssuer) Issue(userID uuid.UUID, role domain.Role) (string, error) {
	now := time.Now()
	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(j.secret)
}

// Parse verifies the signature, the algorithm, and the expiry,
// then returns the user ID and role. Tokens without a role are rejected (fail closed).
func (j *JWTIssuer) Parse(tokenString string) (uuid.UUID, domain.Role, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(
		tokenString,
		&claims,
		func(*jwt.Token) (any, error) { return j.secret, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		return uuid.Nil, "", err
	}
	if claims.Role == "" {
		return uuid.Nil, "", errors.New("token has no role claim")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, "", err
	}
	return userID, claims.Role, nil
}
