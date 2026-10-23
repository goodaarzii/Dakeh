package review

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/middleware"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
)

type Service interface {
	ListByProductID(ctx context.Context, productID uuid.UUID, page, pageSize int) ([]domain.Review, int, bool, error)
	GetMineForProduct(ctx context.Context, userID, productID uuid.UUID) (domain.Review, error)
	Create(ctx context.Context, userID, productID uuid.UUID, rating int, title, body string) (domain.Review, error)
	UpdateMine(ctx context.Context, userID, reviewID uuid.UUID, rating int, title, body string) (domain.Review, error)
	DeleteMine(ctx context.Context, userID, reviewID uuid.UUID) error
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterReviewRoutes keeps catalog review reads public and protects all writes
// and the signed-in user's own-review endpoint.
//
//	GET    /api/v1/products/{product_id}/reviews  (public list)
//	POST   /api/v1/products/{product_id}/reviews  (authenticated create)
//	GET    /api/v1/reviews/mine?product_id=...   (authenticated own review)
//	PATCH  /api/v1/reviews/:review_id            (authenticated owner update)
//	DELETE /api/v1/reviews/:review_id            (authenticated owner delete)
func (h *Handler) RegisterReviewRoutes(api, protected *gin.RouterGroup) {
	// The catalog already registers GET /products/:slug. Reuse that wildcard name here so Gin
	// can share the route tree; for review endpoints, the value is a product UUID, not a slug.
	api.GET("/products/:slug/reviews", h.ListByProduct)
	protected.POST("/products/:slug/reviews", h.Create)
	protected.GET("/reviews/mine", h.GetMineForProduct)
	protected.PATCH("/reviews/:review_id", h.UpdateMine)
	protected.DELETE("/reviews/:review_id", h.DeleteMine)
}

func (h *Handler) ListByProduct(c *gin.Context) {
	productID, ok := parsePathUUID(c, "slug", "invalid product id")
	if !ok {
		return
	}
	var q listQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_query", "invalid review pagination")
		return
	}
	page, pageSize := q.Page, q.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}

	items, count, hasMore, err := h.svc.ListByProductID(c.Request.Context(), productID, page, pageSize)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, reviewListResponse{
		Items: toReviewResponses(items), Count: count, Page: page, PageSize: pageSize, HasMore: hasMore,
	})
}

func (h *Handler) GetMineForProduct(c *gin.Context) {
	userID := authenticatedUserID(c)
	if userID == uuid.Nil {
		return
	}
	productID, err := uuid.Parse(c.Query("product_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product id")
		return
	}
	item, err := h.svc.GetMineForProduct(c.Request.Context(), userID, productID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toReviewResponse(item))
}

func (h *Handler) Create(c *gin.Context) {
	userID := authenticatedUserID(c)
	if userID == uuid.Nil {
		return
	}
	productID, ok := parsePathUUID(c, "slug", "invalid product id")
	if !ok {
		return
	}
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeInvalidReview(c)
		return
	}
	item, err := h.svc.Create(c.Request.Context(), userID, productID, req.Rating, req.Title, req.Body)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, toReviewResponse(item))
}

func (h *Handler) UpdateMine(c *gin.Context) {
	userID := authenticatedUserID(c)
	if userID == uuid.Nil {
		return
	}
	reviewID, ok := parsePathUUID(c, "review_id", "invalid review id")
	if !ok {
		return
	}
	var req reviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeInvalidReview(c)
		return
	}
	item, err := h.svc.UpdateMine(c.Request.Context(), userID, reviewID, req.Rating, req.Title, req.Body)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toReviewResponse(item))
}

func (h *Handler) DeleteMine(c *gin.Context) {
	userID := authenticatedUserID(c)
	if userID == uuid.Nil {
		return
	}
	reviewID, ok := parsePathUUID(c, "review_id", "invalid review id")
	if !ok {
		return
	}
	if err := h.svc.DeleteMine(c.Request.Context(), userID, reviewID); err != nil {
		response.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func authenticatedUserID(c *gin.Context) uuid.UUID {
	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "authenticated user ID is missing")
		return uuid.Nil
	}
	return userID
}

func parsePathUUID(c *gin.Context, param, message string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", message)
		return uuid.Nil, false
	}
	return id, true
}

func writeInvalidReview(c *gin.Context) {
	response.Error(c, http.StatusBadRequest, "invalid_input", "rating must be 1-5 and review body must be 3-5000 characters")
}
