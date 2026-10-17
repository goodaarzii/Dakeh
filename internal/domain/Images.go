package domain

import (
	"time"

	"github.com/google/uuid"
)

// Business rules about image.
const (
	MaxImagesPerProduct = 10
	MaxImageBytes       = 5 << 20 // 5 MB
)

// ImageExtension returns the file extension for an allowed image type.
// Only formats a browser can show and that cannot carry scripts. SVG is NOT allowed on purpose.
func ImageExtension(contentType string) (ext string, ok bool) {
	switch contentType {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	}
	return "", false
}

type ProductImage struct {
	ID         uuid.UUID
	ProductID  uuid.UUID
	VariantID  *uuid.UUID // optional: a photo of one specific variant
	StorageKey string     // name of the file in the bucket
	URL        string     // public address
	AltText    string
	SortOrder  int
	IsPrimary  bool
	CreatedAt  time.Time
}
