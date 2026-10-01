package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(c *gin.Context, status int, data any) {
	c.JSON(status, data)
}

// Error writes the standard error format from docs/API.md:
// { "error": { "code": "...", "message": "..." } }
func Error(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, errorBody{Error: errorDetail{Code: code, Message: message}})
}

func GetTokenFromRequest(r *http.Request) string {
	tokenAuth := r.Header.Get("Authorization")
	tokenQuery := r.URL.Query().Get("token")

	if tokenAuth != "" {
		return tokenAuth
	}

	if tokenQuery != "" {
		return tokenQuery
	}

	return ""
}
