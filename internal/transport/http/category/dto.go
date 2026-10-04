package category

import (
	"time"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// categoryRequest is used for both create (POST) and full update (PUT).
// IsActive is a pointer so that "not sent" (defaults to true) differs from "false".
// Slug format and parent rules (parent exists, no cycles) are checked in the usecase.
type categoryRequest struct {
	ParentID    *uuid.UUID `json:"parent_id"`
	Name        string     `json:"name" binding:"required,max=100"`
	Slug        string     `json:"slug" binding:"required,max=120"`
	Description string     `json:"description" binding:"max=5000"`
	ImageURL    string     `json:"image_url" binding:"omitempty,url,max=500"`
	SortOrder   int        `json:"sort_order" binding:"gte=0"`
	IsActive    *bool      `json:"is_active"`
}

// toDomain leaves ID empty. On update, the handler sets it from the URL.
func (r categoryRequest) toDomain() domain.Category {
	active := true
	if r.IsActive != nil {
		active = *r.IsActive
	}
	return domain.Category{
		ParentID:    r.ParentID,
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		ImageURL:    r.ImageURL,
		SortOrder:   r.SortOrder,
		IsActive:    active,
	}
}

type categoryResponse struct {
	ID          uuid.UUID  `json:"id"`
	ParentID    *uuid.UUID `json:"parent_id"`
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Description string     `json:"description"`
	ImageURL    string     `json:"image_url"`
	SortOrder   int        `json:"sort_order"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// categoryTreeResponse is a category with its children, for menus.
type categoryTreeResponse struct {
	categoryResponse
	Children []categoryTreeResponse `json:"children"`
}

func toCategoryResponse(c domain.Category) categoryResponse {
	return categoryResponse{
		ID:          c.ID,
		ParentID:    c.ParentID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		ImageURL:    c.ImageURL,
		SortOrder:   c.SortOrder,
		IsActive:    c.IsActive,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}

func toCategoryResponses(list []domain.Category) []categoryResponse {
	out := make([]categoryResponse, len(list))
	for i, c := range list {
		out[i] = toCategoryResponse(c)
	}
	return out
}

// toCategoryTreeResponse converts the tree recursively.
// make() (and not nil) makes empty children serialize as [] and never null.
func toCategoryTreeResponse(nodes []domain.CategoryNode) []categoryTreeResponse {
	out := make([]categoryTreeResponse, len(nodes))
	for i, n := range nodes {
		out[i] = categoryTreeResponse{
			categoryResponse: toCategoryResponse(n.Category),
			Children:         toCategoryTreeResponse(n.Children),
		}
	}
	return out
}
