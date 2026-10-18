package address

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/middleware"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, address domain.Address) (domain.Address, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (domain.Address, error)
	Update(ctx context.Context, u domain.Address) (domain.Address, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (int64, error)
	GetUserAddresses(ctx context.Context, userID uuid.UUID) ([]domain.Address, error)
}

type Handler struct {
	svc         Service
	tokenIssuer TokenIssuer
}

type TokenIssuer interface {
	ParseToken(token string) (uuid.UUID, error)
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterAddressRoutes(r *gin.RouterGroup) {

	g := r.Group("/users/address")
	g.POST("", h.Create)
	g.GET("", h.GetUserAddresses)
	g.GET("/:id", h.GetByID)
	g.DELETE("/:id", h.DeleteAddress)
	g.PUT("/:id", h.Update)
}

func (h *Handler) Create(c *gin.Context) {
	var req addressRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		return
	}

	addr, err := h.svc.Create(c.Request.Context(), domain.Address{
		UserID:        middleware.UserID(c),
		Label:         req.Label,
		RecipientName: req.RecipientName,
		Phone:         req.Phone,
		Country:       req.Country,
		Province:      req.Province,
		City:          req.City,
		Street:        req.Street,
		PostalCode:    req.PostalCode,
		Lat:           req.Lat,
		Lng:           req.Lng,
		IsDefault:     req.IsDefault,
	})
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, toAddressResponse(addr))

}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "address not found")
		return
	}
	addr, err := h.svc.GetByID(c.Request.Context(), id, middleware.UserID(c))

	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toAddressResponse(addr))

}

func (h *Handler) GetUserAddresses(c *gin.Context) {

	addresses, err := h.svc.GetUserAddresses(c.Request.Context(), middleware.UserID(c))

	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, toAddressResponses(addresses))

}

func (h *Handler) DeleteAddress(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "address not found")
		return
	}
	deleteAddress, err := h.svc.Delete(c.Request.Context(), id, middleware.UserID(c))
	if err != nil {
		return
	}
	status := http.StatusOK
	if deleteAddress == 0 {
		status = http.StatusNotFound
	}
	response.JSON(c, status, response.NewStatusResponse(deleteAddress))

}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "address not found")
		return
	}
	var req addressRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		return
	}

	addr, err := h.svc.Update(c.Request.Context(),
		domain.Address{
			ID:            id,
			UserID:        middleware.UserID(c),
			Label:         req.Label,
			RecipientName: req.RecipientName,
			Phone:         req.Phone,
			Country:       req.Country,
			Province:      req.Province,
			City:          req.City,
			Street:        req.Street,
			PostalCode:    req.PostalCode,
			Lat:           req.Lat,
			Lng:           req.Lng,
			IsDefault:     req.IsDefault,
		})
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, toAddressResponse(addr))

}
