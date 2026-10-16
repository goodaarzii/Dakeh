package security

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// NewRefreshToken creates a random token for the client and the hash to store in the database.
// 32 random bytes = 256 bits: impossible to guess.
func NewRefreshToken() (plain, hash string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	plain = base64.RawURLEncoding.EncodeToString(b)
	return plain, HashRefreshToken(plain), nil
}

// HashRefreshToken is what we store and look up by.
// SHA-256 (fast) is right here because the token is long and random.
// bcrypt (slow) is for passwords that humans choose, and it could not be used for lookup anyway.
func HashRefreshToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}

// The two methods below let *JWTIssuer be the ONE token dependency of the user usecase:
// it issues access tokens (jwt.go) and creates/hashes refresh tokens (this file).

func (j *JWTIssuer) NewRefreshToken() (plain, hash string, err error) {
	return NewRefreshToken()
}

func (j *JWTIssuer) HashRefreshToken(plain string) string {
	return HashRefreshToken(plain)
}
