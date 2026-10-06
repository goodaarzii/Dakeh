package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
)

const (
	userIDKey = "userID"
	roleKey   = "role"
)

// TokenParser is what the middleware needs. *security.JWTIssuer satisfies it.
type TokenParser interface {
	Parse(token string) (uuid.UUID, domain.Role, error)
}

// Auth answers "who are you?". It verifies the token ONCE and stores
// the user ID and role in the request context. Failure means 401.
func Auth(tokens TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "missing or malformed Authorization header")
			return
		}

		userID, role, err := tokens.Parse(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
			return
		}

		c.Set(userIDKey, userID)
		c.Set(roleKey, role)
		c.Next()
	}
}

// UserID returns the logged-in user's ID. Only call it in handlers behind Auth.
func UserID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(userIDKey)
	id, _ := v.(uuid.UUID)
	return id
}

// Role returns the logged-in user's role. Only call it behind Auth.
func Role(c *gin.Context) domain.Role {
	v, _ := c.Get(roleKey)
	r, _ := v.(domain.Role)
	return r
}
