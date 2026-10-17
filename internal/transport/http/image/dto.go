package image

import (
	"time"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// uploadImageForm holds the TEXT fields of the multipart upload.
// The file itself is read with c.FormFile("file").
type uploadImageForm struct {
	AltText   string `form:"alt_text" binding:"max=255"`
	VariantID string `form:"variant_id" binding:"omitempty,uuid"` // optional
	SortOrder int    `form:"sort_order" binding:"gte=0,lte=1000"`
}

// variantID returns nil when no variant was sent. Call it after binding (the format is validated).
func (f uploadImageForm) variantID() *uuid.UUID {
	if f.VariantID == "" {
		return nil
	}
	id, err := uuid.Parse(f.VariantID)
	if err != nil {
		return nil
	}
	return &id
}

// updateImageRequest changes the text and the order. The file itself cannot be replaced: upload a new one.
type updateImageRequest struct {
	AltText   string `json:"alt_text" binding:"max=255"`
	SortOrder int    `json:"sort_order" binding:"gte=0,lte=1000"`
}

type imageResponse struct {
	ID        uuid.UUID  `json:"id"`
	ProductID uuid.UUID  `json:"product_id"`
	VariantID *uuid.UUID `json:"variant_id"`
	URL       string     `json:"url"`
	AltText   string     `json:"alt_text"`
	SortOrder int        `json:"sort_order"`
	IsPrimary bool       `json:"is_primary"`
	CreatedAt time.Time  `json:"created_at"`
}

func toImageResponse(i domain.ProductImage) imageResponse {
	return imageResponse{
		ID:        i.ID,
		ProductID: i.ProductID,
		VariantID: i.VariantID,
		URL:       i.URL,
		AltText:   i.AltText,
		SortOrder: i.SortOrder,
		IsPrimary: i.IsPrimary,
		CreatedAt: i.CreatedAt,
	}
}

// toImageResponses uses make(), so an empty list is [] and never null.
func toImageResponses(list []domain.ProductImage) []imageResponse {
	out := make([]imageResponse, len(list))
	for i, img := range list {
		out[i] = toImageResponse(img)
	}
	return out
}
