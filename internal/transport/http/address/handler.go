package address

import (
	"context"
	"errors"
	"log/slog"
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
	UpdateAddress(ctx context.Context, id uuid.UUID, userID uuid.UUID) (domain.Address, error)
	DeleteAddress(ctx context.Context, id uuid.UUID, userID uuid.UUID) (int64, error)
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
	auth := r.Group("/user")
	auth.POST("/createAddress", h.Create)
	auth.POST("/getAddressById/{id}", h.GetByID)
	auth.GET("/getUserAddress", h.GetUserAddresses)
	auth.GET("/deleteAddress/{id}", h.GetUserAddresses)
}

func (h *Handler) Create(c *gin.Context) {
	var req addressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
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
		h.writeError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, toAddressResponse(addr))

}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid address id")
		return
	}
	addr, err := h.svc.GetByID(c.Request.Context(), id, middleware.UserID(c))

	if err != nil {
		h.writeError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toAddressResponse(addr))

}

func (h *Handler) GetUserAddresses(c *gin.Context) {

	addresses, err := h.svc.GetUserAddresses(c.Request.Context(), middleware.UserID(c))

	if err != nil {
		h.writeError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, toAddressResponses(addresses))

}

func (h *Handler) DeleteAddress(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid address id")
		return
	}
	deleteAddress, err := h.svc.DeleteAddress(c.Request.Context(), id, middleware.UserID(c))
	if err != nil {
		return
	}
	response.JSON(c, http.StatusOK, NewStatusResponse(deleteAddress))

}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		response.Error(c, http.StatusConflict, "email_taken", "email already registered")
	case errors.Is(err, domain.ErrPhoneTaken):
		response.Error(c, http.StatusConflict, "phone_taken", "phone already registered")
	case errors.Is(err, domain.ErrNotFound):
		response.Error(c, http.StatusNotFound, "not_found", "address not found")
	default:
		slog.Error("user request failed", "err", err)
		response.Error(c, http.StatusInternalServerError, "internal", "something went wrong")
	}
}
