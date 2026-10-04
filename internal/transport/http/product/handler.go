package product

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, address domain.Product) (domain.Product, error)
}

type Handler struct {
	svc         Service
	tokenIssuer TokenIssuer
}

type TokenIssuer interface {
	ParseToken(token string) (uuid.UUID, error)
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterProductRoutes(r *gin.RouterGroup) {

	g := r.Group("/admin/product")
	g.POST("", h.Create)
	//g.GET("", h.GetUserAddresses)
	//g.GET("/:id", h.GetByID)
	//g.DELETE("/:id", h.DeleteAddress)
	//g.PUT("/:id", h.Update)
}

func (h *Handler) Create(c *gin.Context) {

	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	product, err := h.svc.Create(c.Request.Context(), domain.Product{
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Brand:       req.Brand,
		Status:      domain.ProductStatus(req.Status),
	})
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, toProductResponse(product))

}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		response.Error(c, http.StatusConflict, "email_taken", "email already registered")
	case errors.Is(err, domain.ErrPhoneTaken):
		response.Error(c, http.StatusConflict, "phone_taken", "phone already registered")
	case errors.Is(err, domain.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not_found", "address not found")
	default:
		slog.Error("user request failed", "err", err)
		response.Error(c, http.StatusInternalServerError, "internal", "something went wrong")
	}
}
