package catalog

import (
	"strings"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// ---------- request: GET /products?category=shoes&tag=sale&q=run&sort=price_asc&page=2&page_size=20 ----------

type listQuery struct {
	Page     int    `form:"page" binding:"omitempty,gte=1,lte=10000"`
	PageSize int    `form:"page_size" binding:"omitempty,gte=1,lte=50"`
	Category string `form:"category" binding:"omitempty,max=120"` // category slug
	Tag      string `form:"tag" binding:"omitempty,max=60"`       // tag slug
	Q        string `form:"q" binding:"omitempty,max=100"`        // search in the product name
	Sort     string `form:"sort" binding:"omitempty,oneof=newest price_asc price_desc"`
}

// filter applies the defaults: page 1, 20 per page, newest first.
func (q listQuery) filter() domain.ProductFilter {
	page, size := q.Page, q.PageSize
	if page == 0 {
		page = 1
	}
	if size == 0 {
		size = 20
	}
	sort := domain.ProductSort(q.Sort)
	if q.Sort == "" {
		sort = domain.SortNewest
	}
	return domain.ProductFilter{
		CategorySlug: q.Category,
		TagSlug:      q.Tag,
		Search:       strings.TrimSpace(q.Q),
		Sort:         sort,
		Page:         page,
		PageSize:     size,
	}
}

// ---------- categories ----------

// categoryNodeResponse is a category with its children, for menus.
// No timestamps, is_active or sort_order: customers do not need them.
type categoryNodeResponse struct {
	ID       uuid.UUID              `json:"id"`
	Name     string                 `json:"name"`
	Slug     string                 `json:"slug"`
	ImageURL string                 `json:"image_url"`
	Children []categoryNodeResponse `json:"children"`
}

// toCategoryTree converts recursively. make() makes empty children serialize as [] and never null.
func toCategoryTree(nodes []domain.CategoryNode) []categoryNodeResponse {
	out := make([]categoryNodeResponse, len(nodes))
	for i, n := range nodes {
		out[i] = categoryNodeResponse{
			ID:       n.ID,
			Name:     n.Name,
			Slug:     n.Slug,
			ImageURL: n.ImageURL,
			Children: toCategoryTree(n.Children),
		}
	}
	return out
}

// ---------- tags ----------

type tagResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

func toTagResponses(list []domain.Tag) []tagResponse {
	out := make([]tagResponse, len(list))
	for i, t := range list {
		out[i] = tagResponse{ID: t.ID, Name: t.Name, Slug: t.Slug}
	}
	return out
}

// ---------- product list (cards) ----------

// productCardResponse: prices are in USD cents. The frontend formats them as dollars.
// An image_url field will be added when product images exist.
type productCardResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Brand       string    `json:"brand"`
	MinPrice    int64     `json:"min_price"` // "from $X"
	InStock     bool      `json:"in_stock"`
	AvgRating   float64   `json:"avg_rating"`
	RatingCount int       `json:"rating_count"`
}

type productListResponse struct {
	Items    []productCardResponse `json:"items"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
	HasMore  bool                  `json:"has_more"` // is there a next page?
}

func toProductListResponse(cards []domain.ProductCard, page, pageSize int, hasMore bool) productListResponse {
	items := make([]productCardResponse, len(cards))
	for i, c := range cards {
		items[i] = productCardResponse{
			ID:          c.ID,
			Name:        c.Name,
			Slug:        c.Slug,
			Brand:       c.Brand,
			MinPrice:    c.MinPrice,
			InStock:     c.InStock,
			AvgRating:   c.AvgRating,
			RatingCount: c.RatingCount,
		}
	}
	return productListResponse{Items: items, Page: page, PageSize: pageSize, HasMore: hasMore}
}

// ---------- product detail ----------

type categorySummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

// variantResponse is the customer view: in_stock only, never the exact quantities
// (those are business information, and the cart/checkout validate them anyway).
type variantResponse struct {
	ID             uuid.UUID         `json:"id"`
	SKU            string            `json:"sku"`
	Name           string            `json:"name"`
	Price          int64             `json:"price"`            // cents
	CompareAtPrice int64             `json:"compare_at_price"` // cents, 0 = no old price
	Attributes     map[string]string `json:"attributes"`
	InStock        bool              `json:"in_stock"`
}

type productDetailResponse struct {
	ID          uuid.UUID         `json:"id"`
	Name        string            `json:"name"`
	Slug        string            `json:"slug"`
	Description string            `json:"description"`
	Brand       string            `json:"brand"`
	Category    categorySummary   `json:"category"`
	Tags        []tagResponse     `json:"tags"`
	Variants    []variantResponse `json:"variants"`
	AvgRating   float64           `json:"avg_rating"`
	RatingCount int               `json:"rating_count"`
}

func toProductDetailResponse(d domain.ProductDetail) productDetailResponse {
	variants := make([]variantResponse, len(d.Variants))
	for i, v := range d.Variants {
		attrs := v.Variant.Attributes
		if attrs == nil {
			attrs = map[string]string{} // {} and never null
		}
		variants[i] = variantResponse{
			ID:             v.Variant.ID,
			SKU:            v.Variant.SKU,
			Name:           v.Variant.Name,
			Price:          v.Variant.Price,
			CompareAtPrice: v.Variant.CompareAtPrice,
			Attributes:     attrs,
			InStock:        v.Stock.Available() > 0,
		}
	}

	return productDetailResponse{
		ID:          d.Product.ID,
		Name:        d.Product.Name,
		Slug:        d.Product.Slug,
		Description: d.Product.Description,
		Brand:       d.Product.Brand,
		Category:    categorySummary{ID: d.Category.ID, Name: d.Category.Name, Slug: d.Category.Slug},
		Tags:        toTagResponses(d.Tags),
		Variants:    variants,
		AvgRating:   d.Product.AvgRating,
		RatingCount: d.Product.RatingCount,
	}
}
