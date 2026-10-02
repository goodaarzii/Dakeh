package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/transport/http/response"
)

const userIDKey = "userID"

// TokenParser is what the middleware needs. *security.JWTIssuer satisfies it.
type TokenParser interface {
	Parse(token string) (uuid.UUID, error)
}

// Auth reads "Authorization: Bearer <token>", verifies it ONCE,
// and stores the user ID in the request context for handlers to use.
func Auth(tokens TokenParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		scheme, token, ok := strings.Cut(c.GetHeader("Authorization"), " ")
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "missing or malformed Authorization header")
			return
		}

		userID, err := tokens.Parse(token)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "unauthorized", "invalid or expired token")
			return
		}

		c.Set(userIDKey, userID)
		c.Next()
	}
}

// UserID returns the logged-in user's ID. Only call it in handlers behind Auth.
func UserID(c *gin.Context) uuid.UUID {
	v, _ := c.Get(userIDKey)
	id, _ := v.(uuid.UUID)
	return id
}
