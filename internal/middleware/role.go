package middleware

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
)

// RequireRole answers "are you allowed?". It must run AFTER Auth.
// A missing or unknown role is not in the list, so it fails closed with 403.
func RequireRole(allowed ...domain.Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !slices.Contains(allowed, Role(c)) {
			response.Error(c, http.StatusForbidden, "forbidden", "you do not have permission to do this")
			return
		}
		c.Next()
	}
}
