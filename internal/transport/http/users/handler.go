package user

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/goodaarzii/dakeh/internal/transport/http/response"
	userusecase "github.com/goodaarzii/dakeh/internal/usecase/user"
)

// Service is what the handler needs from the usecase layer.
type Service interface {
	Register(ctx context.Context, in userusecase.RegisterInput) (userusecase.AuthResult, error)
	Login(ctx context.Context, email, password string) (userusecase.AuthResult, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterUserRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/login", h.refresh)
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
		response.WriteError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, toAuthResponse(res))
}

func (h *Handler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toAuthResponse(res))

}
