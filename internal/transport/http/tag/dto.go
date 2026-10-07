package tag

import (
	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// tagRequest is used for both create (POST) and full update (PUT).
// Slug format (lowercase letters, digits, dashes) is checked in the usecase.
type tagRequest struct {
	Name string `json:"name" binding:"required,max=50"`
	Slug string `json:"slug" binding:"required,max=60"`
}

// toDomain leaves ID empty. On update, the handler sets it from the URL.
func (r tagRequest) toDomain() domain.Tag {
	return domain.Tag{Name: r.Name, Slug: r.Slug}
}

// tagResponse is deliberately small, so it can also be embedded inside product responses.
type tagResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Slug string    `json:"slug"`
}

func toTagResponse(t domain.Tag) tagResponse {
	return tagResponse{ID: t.ID, Name: t.Name, Slug: t.Slug}
}

// toTagResponses uses make(), so an empty list becomes [] and never null.
func toTagResponses(list []domain.Tag) []tagResponse {
	out := make([]tagResponse, len(list))
	for i, t := range list {
		out[i] = toTagResponse(t)
	}
	return out
}

type AddTagToProductRequest struct {
	ProductID uuid.UUID `form:"product_id"`
	TagID     uuid.UUID `form:"tag_id"`
}

type setProductTagsRequest struct {
	TagIDs []uuid.UUID `json:"tag_ids" binding:"required,max=50,dive,required"`
}
