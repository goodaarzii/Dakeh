package address

import (
	"time"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// addressRequest is used for both create (POST) and full update (PUT).
// Lat/Lng are pointers so that a real 0 is accepted and a missing field is rejected:
// with a plain float64, `required` treats 0 as "missing".
type addressRequest struct {
	Label         string   `json:"label" binding:"max=50"`
	RecipientName string   `json:"recipient_name" binding:"required,max=255"`
	Phone         string   `json:"phone" binding:"required,min=7,max=20"`
	Country       string   `json:"country" binding:"required,max=100"`
	Province      string   `json:"province" binding:"required,max=100"`
	City          string   `json:"city" binding:"required,max=100"`
	Street        string   `json:"street" binding:"required,max=500"`
	PostalCode    string   `json:"postal_code" binding:"max=20"`
	Lat           *float64 `json:"lat" binding:"required,gte=-90,lte=90"`
	Lng           *float64 `json:"lng" binding:"required,gte=-180,lte=180"`
	IsDefault     bool     `json:"is_default"` // only used on create
}

// addressResponse is what the client receives. No user_id, the client already knows who it is.
type addressResponse struct {
	ID            uuid.UUID `json:"id"`
	Label         string    `json:"label"`
	RecipientName string    `json:"recipient_name"`
	Phone         string    `json:"phone"`
	Country       string    `json:"country"`
	Province      string    `json:"province"`
	City          string    `json:"city"`
	Street        string    `json:"street"`
	PostalCode    string    `json:"postal_code"`
	Lat           float64   `json:"lat"`
	Lng           float64   `json:"lng"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func toAddressResponse(a domain.Address) addressResponse {
	return addressResponse{
		ID:            a.ID,
		Label:         a.Label,
		RecipientName: a.RecipientName,
		Phone:         a.Phone,
		Country:       a.Country,
		Province:      a.Province,
		City:          a.City,
		Street:        a.Street,
		PostalCode:    a.PostalCode,
		Lat:           a.Lat,
		Lng:           a.Lng,
		IsDefault:     a.IsDefault,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}
