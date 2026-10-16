package catalog

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
)

type Service interface {
	// CategoryTree Active categories as a tree (the usecase builds it from the flat list).
	CategoryTree(ctx context.Context) ([]domain.CategoryNode, error)

	GetTags(ctx context.Context) ([]domain.Tag, error)

	// ListProducts One page of product cards. hasMore is true when a next page exists
	// (the usecase asks the repo for PageSize+1 rows and cuts the extra one).
	ListProducts(ctx context.Context, f domain.ProductFilter) (cards []domain.ProductCard, hasMore bool, err error)

	// GetActiveProductBySlug The whole product page. Unknown, draft or variant-less product -> domain.ErrNotFound.
	GetActiveProductBySlug(ctx context.Context, slug string) (domain.ProductDetail, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterCatalogRoutes registers the PUBLIC shop endpoints (no login).
// Pass the plain api group in routes.go, NOT the protected or admin group.
//
//	GET /api/v1/categories
//	GET /api/v1/tags
//	GET /api/v1/products?category=&tag=&q=&sort=&page=&page_size=
//	GET /api/v1/products/:slug
func (h *Handler) RegisterCatalogRoutes(r *gin.RouterGroup) {
	r.GET("/categories", h.ListCategories)
	r.GET("/tags", h.ListTags)
	r.GET("/products", h.ListProducts)
	r.GET("/products/:slug", h.GetProduct)
}

func (h *Handler) ListCategories(c *gin.Context) {
	tree, err := h.svc.CategoryTree(c.Request.Context())
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toCategoryTree(tree))
}

func (h *Handler) ListTags(c *gin.Context) {
	tags, err := h.svc.GetTags(c.Request.Context())
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toTagResponses(tags))
}

func (h *Handler) ListProducts(c *gin.Context) {
	var q listQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	f := q.filter() // applies the defaults: page 1, 20 per page, newest first

	cards, hasMore, err := h.svc.ListProducts(c.Request.Context(), f)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toProductListResponse(cards, f.Page, f.PageSize, hasMore))
}

func (h *Handler) GetProduct(c *gin.Context) {
	detail, err := h.svc.GetActiveProductBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		response.WriteError(c, err) // ErrNotFound -> 404
		return
	}
	response.JSON(c, http.StatusOK, toProductDetailResponse(detail))
}
