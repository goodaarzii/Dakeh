package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORS lets a browser frontend on another origin (for example http://localhost:5173)
// call this API. Without it the browser blocks every request, even if the API works in Postman.
// allowedOrigins is an exact list, never "*" together with Authorization headers.
// Register it FIRST: r.Use(middleware.CORS(origins)), before the route groups.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(allowedOrigins))
	for _, o := range allowedOrigins {
		allowed[o] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && allowed[origin] {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Vary", "Origin")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			h.Set("Access-Control-Max-Age", "600")
		}

		// The browser first sends an OPTIONS "preflight" request. Answer it here and stop.
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
