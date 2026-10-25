package order

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/middleware"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
)

type Service interface {
	Create(ctx context.Context, userID, addressID uuid.UUID, idempotencyKey string) (domain.Order, bool, error)
	ListMine(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]domain.Order, int, bool, error)
	CancelMine(ctx context.Context, userID, orderID uuid.UUID) (domain.Order, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler { return &Handler{svc: svc} }

// RegisterOrderRoutes mounts authenticated order history and order submission.
// POST /orders creates a submitted (not paid) order; POST /orders/:order_id/cancel cancels it.
func (h *Handler) RegisterOrderRoutes(r *gin.RouterGroup) {
	g := r.Group("/orders")
	g.POST("", h.Create)
	g.GET("", h.ListMine)
	g.POST("/:order_id/cancel", h.CancelMine)
}

func (h *Handler) Create(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "authenticated user ID is missing")
		return
	}
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.AddressID == uuid.Nil {
		response.Error(c, http.StatusBadRequest, "invalid_input", "a valid address_id is required")
		return
	}
	idempotencyKey := c.GetHeader("Idempotency-Key")
	if strings.TrimSpace(idempotencyKey) == "" || len(idempotencyKey) > 128 {
		response.Error(c, http.StatusBadRequest, "idempotency_key_required", "a non-empty Idempotency-Key header (up to 128 characters) is required")
		return
	}
	created, cartCleared, err := h.svc.Create(c.Request.Context(), userID, req.AddressID, idempotencyKey)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusCreated, createOrderResponse{Order: toOrderResponse(created), CartCleared: cartCleared})
}

func (h *Handler) ListMine(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "authenticated user ID is missing")
		return
	}
	var q listQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_query", "invalid order pagination")
		return
	}
	page, pageSize := q.Page, q.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 10
	}
	items, count, hasMore, err := h.svc.ListMine(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	out := make([]orderResponse, len(items))
	for i, item := range items {
		out[i] = toOrderResponse(item)
	}
	response.JSON(c, http.StatusOK, orderListResponse{Items: out, Count: count, Page: page, PageSize: pageSize, HasMore: hasMore})
}

func (h *Handler) CancelMine(c *gin.Context) {
	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		response.Error(c, http.StatusUnauthorized, "unauthorized", "authenticated user ID is missing")
		return
	}
	orderID, err := uuid.Parse(c.Param("order_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid order id")
		return
	}
	item, err := h.svc.CancelMine(c.Request.Context(), userID, orderID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toOrderResponse(item))
}
