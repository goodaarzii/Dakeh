package favorite

import (
	"time"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/google/uuid"
)

type favoriteResponse struct {
	ProductID   uuid.UUID `json:"product_id"`
	FavoritedAt time.Time `json:"favorited_at"`
	IsFavorite  bool      `json:"is_favorite"`
}

type favoriteProductResponse struct {
	ID          uuid.UUID `json:"id"`
	CategoryID  uuid.UUID `json:"category_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Brand       string    `json:"brand"`
	MinPrice    int64     `json:"min_price"`
	InStock     bool      `json:"in_stock"`
	ImageURL    string    `json:"image_url"`
	AvgRating   float64   `json:"avg_rating"`
	RatingCount int       `json:"rating_count"`
	FavoritedAt time.Time `json:"favorited_at"`
}

type favoritesListResponse struct {
	Items []favoriteProductResponse `json:"items"`
	Count int                       `json:"count"`
}

func toFavoriteResponse(f domain.Favorite) favoriteResponse {
	return favoriteResponse{ProductID: f.ProductID, FavoritedAt: f.CreatedAt, IsFavorite: true}
}

func toFavoritesListResponse(favorites []domain.FavoriteProduct) favoritesListResponse {
	items := make([]favoriteProductResponse, len(favorites))
	for i, favorite := range favorites {
		p := favorite.Product
		items[i] = favoriteProductResponse{
			ID:          p.ID,
			CategoryID:  p.CategoryID,
			Name:        p.Name,
			Slug:        p.Slug,
			Brand:       p.Brand,
			MinPrice:    p.MinPrice,
			InStock:     p.InStock,
			ImageURL:    p.ImageURL,
			AvgRating:   p.AvgRating,
			RatingCount: p.RatingCount,
			FavoritedAt: favorite.CreatedAt,
		}
	}
	return favoritesListResponse{Items: items, Count: len(items)}
}
