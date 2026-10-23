package review

import (
	"time"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

type reviewRequest struct {
	Rating int    `json:"rating" binding:"required,gte=1,lte=5"`
	Title  string `json:"title" binding:"omitempty,max=120"`
	Body   string `json:"body" binding:"required,min=3,max=5000"`
}

type reviewResponse struct {
	ID           uuid.UUID `json:"id"`
	ProductID    uuid.UUID `json:"product_id"`
	ReviewerName string    `json:"reviewer_name"`
	Rating       int       `json:"rating"`
	Title        string    `json:"title"`
	Body         string    `json:"body"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type reviewListResponse struct {
	Items    []reviewResponse `json:"items"`
	Count    int              `json:"count"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
	HasMore  bool             `json:"has_more"`
}

type listQuery struct {
	Page     int `form:"page" binding:"omitempty,gte=1"`
	PageSize int `form:"page_size" binding:"omitempty,gte=1,lte=50"`
}

func toReviewResponse(r domain.Review) reviewResponse {
	return reviewResponse{
		ID: r.ID, ProductID: r.ProductID, ReviewerName: r.ReviewerName,
		Rating: r.Rating, Title: r.Title, Body: r.Body,
		CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
	}
}

func toReviewResponses(reviews []domain.Review) []reviewResponse {
	out := make([]reviewResponse, len(reviews))
	for i, r := range reviews {
		out[i] = toReviewResponse(r)
	}
	return out
}
