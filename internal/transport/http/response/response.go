package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goodaarzii/dakeh/internal/domain"
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

func WriteError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		Error(c, http.StatusConflict, "email_taken", "email already registered")
	case errors.Is(err, domain.ErrPhoneTaken):
		Error(c, http.StatusConflict, "phone_taken", "phone already registered")
	case errors.Is(err, domain.ErrNotFound):
		Error(c, http.StatusNotFound, "not_found", "address not found")
	default:
		Error(c, http.StatusInternalServerError, "internal", "something went wrong")
	}
}

type StatusResponse struct {
	Status    int64  `json:"code"` // rows changed: 0 or 1
	Message   string `json:"message"`
	IsSuccess bool   `json:"is_success"`
}

func NewStatusResponse(rows int64) StatusResponse {
	if rows > 0 {
		return StatusResponse{Status: rows, Message: "success", IsSuccess: true}
	}
	return StatusResponse{Status: rows, Message: "failed", IsSuccess: false}

}
