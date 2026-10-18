package response

import (
	"errors"
	"log/slog"
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

func WriteError(c *gin.Context, err error) {
	// Our own validation message (written by us in the usecase, never from the database) is safe to show.
	if errors.Is(err, domain.ErrInvalidInput) {
		Error(c, http.StatusBadRequest, "invalid_input", err.Error())
		return
	}
	for _, e := range errorMap {
		if errors.Is(err, e.err) {
			Error(c, e.status, e.code, e.msg)
			return
		}
	}

	// Unknown error: the client gets a generic 500, and the REAL error goes to the log.
	// Without this line you would never find out why a 500 happened.
	slog.Error("unhandled error", "method", c.Request.Method, "path", c.Request.URL.Path, "err", err)
	Error(c, http.StatusInternalServerError, "internal", "something went wrong")
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

// Every domain error has one HTTP status, one snake_case code and one FIXED message.
// The codes are the contract with the frontend, so keep them stable.
var errorMap = []struct {
	err    error
	status int
	code   string
	msg    string
}{
	{domain.ErrEmailTaken, http.StatusConflict, "email_taken", "email already registered"},
	{domain.ErrPhoneTaken, http.StatusConflict, "phone_taken", "phone already registered"},
	{domain.ErrSlugTaken, http.StatusConflict, "slug_taken", "slug is already in use"},
	{domain.ErrTagTaken, http.StatusConflict, "tag_taken", "tag name or slug is already in use"},
	{domain.ErrSKUTaken, http.StatusConflict, "sku_taken", "sku is already in use"},
	{domain.ErrCategoryInUse, http.StatusConflict, "category_in_use", "category still has products or child categories"},
	{domain.ErrConflict, http.StatusConflict, "conflict", "this conflicts with existing data"},
	{domain.ErrTooManyImages, http.StatusConflict, "too_many_images", "too many images for this product"},
	{domain.ErrInsufficientStock, http.StatusConflict, "insufficient_stock", "not enough stock"},
	{domain.ErrVariantUnavailable, http.StatusConflict, "variant_unavailable", "this item is not available"},
	{domain.ErrCartFull, http.StatusConflict, "cart_full", "the cart is full"},

	{domain.ErrNotFound, http.StatusNotFound, "not_found", "resource not found"},

	{domain.ErrInvalidCredentials, http.StatusUnauthorized, "invalid_credentials", "invalid email or password"},
	{domain.ErrInvalidRefreshToken, http.StatusUnauthorized, "invalid_refresh_token", "session expired, please log in again"},
	{domain.ErrUnauthorized, http.StatusUnauthorized, "unauthorized", "please log in again"},
	{domain.ErrUserInactive, http.StatusForbidden, "user_inactive", "this account is disabled"},

	{domain.ErrUnsupportedImage, http.StatusUnsupportedMediaType, "unsupported_image_type", "only JPEG, PNG and WebP images are allowed"},
	{domain.ErrImageTooLarge, http.StatusRequestEntityTooLarge, "image_too_large", "the image is too large"},
}
