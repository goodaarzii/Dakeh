package product

import (
	"time"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// productRequest is used for both create (POST) and full update (PUT).
// Status is optional on create and defaults to "draft".
// Slug format and "category exists" are checked in the usecase.
type productRequest struct {
	CategoryID  uuid.UUID `json:"category_id" binding:"required"`
	Name        string    `json:"name" binding:"required,max=255"`
	Slug        string    `json:"slug" binding:"required,max=280"`
	Description string    `json:"description" binding:"max=20000"`
	Brand       string    `json:"brand" binding:"max=100"`
	Status      string    `json:"status" binding:"omitempty,oneof=draft active archived"`
}

// toDomain leaves ID empty. On update, the handler sets it from the URL.
func (r productRequest) toDomain() domain.Product {
	status := domain.ProductStatusDraft
	if r.Status != "" {
		status = domain.ProductStatus(r.Status)
	}
	return domain.Product{
		CategoryID:  r.CategoryID,
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		Brand:       r.Brand,
		Status:      status,
	}
}

// listQuery holds the query-string parameters: GET /admin/products?page=2&page_size=20&category_id=...
type listQuery struct {
	Page       int    `form:"page" binding:"omitempty,gte=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,gte=1,lte=100"`
	CategoryID string `form:"category_id" binding:"omitempty,uuid"`
}

// paging applies the defaults: page 1, 20 items per page.
func (q listQuery) paging() (page, pageSize int) {
	page, pageSize = q.Page, q.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	return page, pageSize
}

type productResponse struct {
	ID          uuid.UUID `json:"id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Brand       string    `json:"brand"`
	Status      string    `json:"status"`
	AvgRating   float64   `json:"avg_rating"`
	RatingCount int       `json:"rating_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// Variants and images will be added here when those features exist.
}

type productListResponse struct {
	Items    []productResponse `json:"items"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

func toProductResponse(p domain.Product) productResponse {
	return productResponse{
		ID:          p.ID,
		CategoryID:  p.CategoryID,
		Name:        p.Name,
		Slug:        p.Slug,
		Description: p.Description,
		Brand:       p.Brand,
		Status:      string(p.Status),
		AvgRating:   p.AvgRating,
		RatingCount: p.RatingCount,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func toProductResponses(list []domain.Product) []productResponse {
	out := make([]productResponse, len(list))
	for i, p := range list {
		out[i] = toProductResponse(p)
	}
	return out
}

func toProductListResponse(list []domain.Product, page, pageSize int) productListResponse {
	return productListResponse{
		Items:    toProductResponses(list),
		Page:     page,
		PageSize: pageSize,
	}
}
