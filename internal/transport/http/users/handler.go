package user

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/middleware"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
	userusecase "github.com/goodaarzii/dakeh/internal/usecase/user"
)

// Service is what the handler needs from the usecase layer.
type Service interface {
	Register(ctx context.Context, in userusecase.RegisterInput) (userusecase.AuthResult, error)
	Login(ctx context.Context, in userusecase.LoginInput) (userusecase.AuthResult, error)
	Refresh(ctx context.Context, in userusecase.RefreshInput) (userusecase.AuthResult, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterUserRoutes: PUBLIC routes. Pass the plain api group.
//
//	POST /api/v1/auth/register
//	POST /api/v1/auth/login
//	POST /api/v1/auth/refresh
//	POST /api/v1/auth/logout
func (h *Handler) RegisterUserRoutes(r *gin.RouterGroup) {
	auth := r.Group("/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.Refresh)
	auth.POST("/logout", h.Logout)
}

// RegisterUserProtectedRoutes: routes for a LOGGED-IN user. Pass the protected group.
//
//	POST /api/v1/auth/logout-all
//	GET  /api/v1/users/me
func (h *Handler) RegisterUserProtectedRoutes(r *gin.RouterGroup) {
	r.POST("/auth/logout-all", h.LogoutAll)
	r.GET("/users/me", h.Me)
}

func (h *Handler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	res, err := h.svc.Register(c.Request.Context(), userusecase.RegisterInput{
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  req.Password,
		FullName:  req.FullName,
		UserAgent: c.Request.UserAgent(),
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

	res, err := h.svc.Login(c.Request.Context(), userusecase.LoginInput{
		Email:     req.Email,
		Password:  req.Password,
		UserAgent: c.Request.UserAgent(),
	})
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, toAuthResponse(res))
}

func (h *Handler) Refresh(c *gin.Context) {
	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	res, err := h.svc.Refresh(c.Request.Context(), userusecase.RefreshInput{
		RefreshToken: req.RefreshToken,
		UserAgent:    c.Request.UserAgent(),
	})
	if err != nil {
		response.WriteError(c, err) // ErrInvalidRefreshToken -> 401
		return
	}

	response.JSON(c, http.StatusOK, toTokenResponse(res))
}

func (h *Handler) Logout(c *gin.Context) {
	var req refreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	if err := h.svc.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		response.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent) // always 204, even for an unknown token
}

func (h *Handler) LogoutAll(c *gin.Context) {
	if err := h.svc.LogoutAll(c.Request.Context(), middleware.UserID(c)); err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, response.NewStatusResponse(1))

}

// Me returns the logged-in user. The ID comes from the token (set by the Auth middleware).
func (h *Handler) Me(c *gin.Context) {
	u, err := h.svc.GetByID(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toUserResponse(u))
}
