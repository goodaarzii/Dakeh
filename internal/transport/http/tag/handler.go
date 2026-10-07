package tag

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, address domain.Tag) (domain.Tag, error)
	Update(ctx context.Context, address domain.Tag) (domain.Tag, error)
	Delete(ctx context.Context, id uuid.UUID) (int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Tag, error)
	AddTagToProduct(ctx context.Context, id uuid.UUID, tagID []uuid.UUID) error
	RemoveTagFromProduct(ctx context.Context, id uuid.UUID, tagID uuid.UUID) (int64, error)
	ClearProductTags(ctx context.Context, id uuid.UUID) error
	ListTags(ctx context.Context) ([]domain.Tag, error)
	ListTagsByProduct(ctx context.Context, id uuid.UUID) ([]domain.Tag, error)
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

func (h *Handler) RegisterTagRoutes(r *gin.RouterGroup) {
	g := r.Group("/tag")
	g.POST("", h.Create)
	g.GET("", h.ListTags)
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)

	// tags of a product (the product is the "parent" in the URL)
	r.GET("/product/:id/tags", h.ListTagsByProduct)
	r.PUT("/product/:id/tags", h.SetProductTags) // replace the whole set
	r.DELETE("/product/:id/tags", h.ClearProductTags)
	r.DELETE("/product/tags/remove", h.RemoveTagFromProduct)
}

func (h *Handler) Create(c *gin.Context) {
	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	create, err := h.svc.Create(c.Request.Context(), domain.Tag{
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, toTagResponse(create))
}

func (h *Handler) Update(c *gin.Context) {
	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	create, err := h.svc.Update(c.Request.Context(), domain.Tag{
		Name: req.Name,
		Slug: req.Slug,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.JSON(c, http.StatusCreated, toTagResponse(create))
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid address id")
		return
	}
	deleted, err := h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, response.NewStatusResponse(deleted))
}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid address id")
		return
	}
	byID, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, toTagResponse(byID))
}

func (h *Handler) SetProductTags(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product id")
		return
	}

	var req setProductTagsRequest
	if err := c.ShouldBindJSON(&req); err != nil { // JSON body, not Query
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if err := h.svc.AddTagToProduct(c.Request.Context(), productID, req.TagIDs); err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, response.NewStatusResponse(1))
}

func (h *Handler) RemoveTagFromProduct(c *gin.Context) {
	var q AddTagToProductRequest
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	isRemoved, err := h.svc.RemoveTagFromProduct(c.Request.Context(), q.ProductID, q.TagID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, response.NewStatusResponse(isRemoved))

}

func (h *Handler) ClearProductTags(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid address id")
		return
	}
	err = h.svc.ClearProductTags(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, response.NewStatusResponse(1))
}

func (h *Handler) ListTags(c *gin.Context) {
	tags, err := h.svc.ListTags(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	response.JSON(c, http.StatusOK, toTagResponses(tags))

}

func (h *Handler) ListTagsByProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product id")
		return
	}
	tags, err := h.svc.ListTagsByProduct(c.Request.Context(), productID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, toTagResponses(tags))
}
