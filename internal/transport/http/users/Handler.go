package user

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
	userusecase "github.com/goodaarzii/dakeh/internal/usecase/user"
)

// Service is what the handler needs from the usecase layer.
type Service interface {
	Register(ctx context.Context, in userusecase.RegisterInput) (userusecase.AuthResult, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.POST("/register", h.Register)
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	res, err := h.svc.Register(c.Request.Context(), userusecase.RegisterInput{
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		h.writeError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, toAuthResponse(res))
}

func (h *Handler) writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrEmailTaken):
		response.Error(c, http.StatusConflict, "email_taken", "email already registered")
	case errors.Is(err, domain.ErrPhoneTaken):
		response.Error(c, http.StatusConflict, "phone_taken", "phone already registered")
	default:
		slog.Error("user request failed", "err", err)
		response.Error(c, http.StatusInternalServerError, "internal", "something went wrong")
	}
}
