package user

import (
	"time"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
	userusecase "github.com/goodaarzii/dakeh/internal/usecase/user"
)

type registerRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Phone    string `json:"phone" binding:"required,min=7,max=20"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	FullName string `json:"full_name" binding:"required,max=255"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
}

// refreshTokenRequest is used by POST /auth/refresh and POST /auth/logout.
type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required,max=200"`
}

// userResponse has no password_hash and no token. A user is not a token.
type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// authResponse is returned by register and login.
type authResponse struct {
	User         userResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"` // always "Bearer"
	ExpiresIn    int          `json:"expires_in"` // seconds until the ACCESS token expires
}

// tokenResponse is returned by refresh: tokens only, the client already knows the user.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func toUserResponse(u domain.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Phone:     u.Phone,
		FullName:  u.FullName,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
	}
}

func toAuthResponse(r userusecase.AuthResult) authResponse {
	return authResponse{
		User:         toUserResponse(r.User),
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(r.ExpiresIn.Seconds()),
	}
}

func toTokenResponse(r userusecase.AuthResult) tokenResponse {
	return tokenResponse{
		AccessToken:  r.AccessToken,
		RefreshToken: r.RefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(r.ExpiresIn.Seconds()),
	}
}
