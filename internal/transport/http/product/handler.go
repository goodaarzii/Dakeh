package product

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, address domain.Product) (domain.Product, error)
	Update(ctx context.Context, address domain.Product) (domain.Product, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Product, error)
	Delete(ctx context.Context, id uuid.UUID) (int64, error)
	ListProducts(ctx context.Context, offset int, pageSize int) ([]domain.Product, error)
	ListProductsByCategory(ctx context.Context, categoryID *uuid.UUID, offset int, pageSize int) ([]domain.Product, error)
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

	g := r.Group("/product")
	g.GET("/", h.ListProducts)
	g.POST("", h.Create)
	g.GET("/:id", h.GetByID)
	g.DELETE("/:id", h.Delete)
	g.PUT("/:id", h.Update)
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
		Status:      domain.ProductStatusActive,
	})
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, toProductResponse(product))

}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid address id")
		return
	}
	var req productRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	update, err := h.svc.Update(c.Request.Context(), domain.Product{
		ID:          id,
		CategoryID:  req.CategoryID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Brand:       req.Brand,
		Status:      domain.ProductStatus(req.Status),
	})
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toProductResponse(update))
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid address id")
		return
	}
	product, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toProductResponse(product))
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid address id")

		return
	}

	deleteResult, err := h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		response.WriteError(c, err)
		return

	}

	response.JSON(c, http.StatusOK, response.NewStatusResponse(deleteResult))

}
func (h *Handler) ListProductsByCategory(c *gin.Context) {
	var q listQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	page, pageSize, _ := q.paging()

	var categoryID *uuid.UUID
	if q.CategoryID != "" {
		id, err := uuid.Parse(q.CategoryID)
		if err != nil {
			response.Error(c, http.StatusBadRequest, "invalid_category_id", "invalid category id")
			return
		}
		categoryID = &id
	}

	products, err := h.svc.ListProductsByCategory(c.Request.Context(), categoryID, page, pageSize)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, toProductListResponse(products, page, pageSize))
}
func (h *Handler) ListProducts(c *gin.Context) {
	var q listQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	page, pageSize, offset := q.paging()

	products, err := h.svc.ListProducts(c.Request.Context(), offset, pageSize) // offset first, then size
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toProductListResponse(products, page, pageSize))

}
