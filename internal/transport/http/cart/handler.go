package cart

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
	Get(ctx context.Context, userID uuid.UUID) (domain.Cart, error)
	SetQuantity(ctx context.Context, userID, variantID uuid.UUID, quantity int) (domain.Cart, error)
	RemoveItem(ctx context.Context, userID, variantID uuid.UUID) (domain.Cart, error)
	Clear(ctx context.Context, userID uuid.UUID) error
	Merge(ctx context.Context, userID uuid.UUID, guest []domain.CartItem) (domain.Cart, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterCartRoutes: for a LOGGED-IN user. Pass the protected group.
// The cart belongs to the user in the token (middleware.UserID), never to an ID sent by the client.
//
//	GET    /api/v1/cart
//	DELETE /api/v1/cart
//	PUT    /api/v1/cart/items/:variant_id    {"quantity": 2}
//	DELETE /api/v1/cart/items/:variant_id
//	POST   /api/v1/cart/merge                {"items": [{"variant_id": "...", "quantity": 1}]}
func (h *Handler) RegisterCartRoutes(r *gin.RouterGroup) {
	g := r.Group("/cart")
	g.GET("", h.Get)
	g.DELETE("", h.Clear)
	g.PUT("/items/:variant_id", h.SetQuantity)
	g.DELETE("/items/:variant_id", h.RemoveItem)
	g.POST("/merge", h.Merge)
}

func (h *Handler) Get(c *gin.Context) {
	cart, err := h.svc.Get(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toCartResponse(cart))
}

func (h *Handler) SetQuantity(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("variant_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid variant id")
		return
	}

	var req setQuantityRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		return
	}

	cart, err := h.svc.SetQuantity(c.Request.Context(), middleware.UserID(c), variantID, req.Quantity)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toCartResponse(cart))
}

func (h *Handler) RemoveItem(c *gin.Context) {
	variantID, err := uuid.Parse(c.Param("variant_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid variant id")
		return
	}

	cart, err := h.svc.RemoveItem(c.Request.Context(), middleware.UserID(c), variantID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toCartResponse(cart))
}

func (h *Handler) Clear(c *gin.Context) {
	if err := h.svc.Clear(c.Request.Context(), middleware.UserID(c)); err != nil {
		response.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) Merge(c *gin.Context) {
	var req mergeRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		return
	}

	cart, err := h.svc.Merge(c.Request.Context(), middleware.UserID(c), req.toDomain())
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toCartResponse(cart))
}
