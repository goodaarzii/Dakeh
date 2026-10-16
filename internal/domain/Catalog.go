package domain

import "github.com/google/uuid"

// ProductSort is how a product list is ordered.
type ProductSort string

const (
	SortNewest    ProductSort = "newest"
	SortPriceAsc  ProductSort = "price_asc"
	SortPriceDesc ProductSort = "price_desc"
)

// ProductFilter is what the shop pages ask for. Empty fields mean "no filter".
// The usecase turns CategorySlug into a list of category IDs (the category and all its children).
type ProductFilter struct {
	CategorySlug string
	TagSlug      string
	Search       string
	Sort         ProductSort
	Page         int
	PageSize     int
}

// ProductCard is the small version of a product, for lists.
// MinPrice is the lowest price among its sellable variants, in cents ("from $X").
// ImageURL is the primary image, or "" when the product has no image.
type ProductCard struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	Name        string
	Slug        string
	Brand       string
	MinPrice    int64
	InStock     bool
	ImageURL    string
	AvgRating   float64
	RatingCount int
}

// ProductDetail is everything the product page needs in ONE response.
// Variants reuse VariantWithStock: the public DTO turns the stock into in_stock.
type ProductDetail struct {
	Product  Product
	Category Category
	Variants []VariantWithStock
	Tags     []Tag
	Images   []ProductImage
}

// ProductQuery is what the repository needs for a product list.
// It is built by the usecase from ProductFilter:
// slugs become IDs (CategoryIDs), and page numbers become Limit and Offset.
// Passing one struct (instead of five loose arguments) means you can never mix up two ints or two strings.
type ProductQuery struct {
	CategoryIDs []uuid.UUID // empty = no category filter
	TagSlug     string      // empty = no tag filter
	Search      string      // empty = no search
	Sort        ProductSort
	Limit       int
	Offset      int
}
