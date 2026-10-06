package category

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, address domain.Category) (domain.Category, error)
	Update(ctx context.Context, address domain.Category) (domain.Category, error)
	CategoryList(ctx context.Context) ([]domain.Category, error)
	FindByID(ctx context.Context, id uuid.UUID) (domain.Category, error)
	Delete(ctx context.Context, id uuid.UUID) (int64, error)
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

func (h *Handler) RegisterCategoryRoutes(r *gin.RouterGroup) {

	g := r.Group("/admin/category")
	g.POST("", h.Create)
	g.GET("", h.CategoryList)
	g.GET("/:id", h.FindByID)
	g.DELETE("/:id", h.Delete)
	g.PUT("/:id", h.Update)
}

func (h *Handler) Create(c *gin.Context) {
	//IsActive must be in request
	var req categoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	category, err := h.svc.Create(context.Background(), domain.Category{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		ImageURL:    req.ImageURL,
		SortOrder:   req.SortOrder,
		IsActive:    *req.IsActive,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, toCategoryResponse(category))

}

func (h *Handler) CategoryList(c *gin.Context) {
	list, err := h.svc.CategoryList(c)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, toCategoryResponses(list))
}

func (h *Handler) FindByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	category, err := h.svc.FindByID(c, id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, toCategoryResponse(category))
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	var request categoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	category, err := h.svc.Update(context.Background(), domain.Category{
		ID:          id,
		ParentID:    request.ParentID,
		Name:        request.Name,
		Slug:        request.Slug,
		Description: request.Description,
		ImageURL:    request.ImageURL,
		SortOrder:   request.SortOrder,
		IsActive:    *request.IsActive,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())

		return
	}
	response.JSON(c, http.StatusOK, toCategoryResponse(category))
}
func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	deleted, err := h.svc.Delete(context.Background(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, response.NewStatusResponse(deleted))

}
