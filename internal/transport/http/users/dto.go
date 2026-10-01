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

// userResponse has no password_hash and no token. A user is not a token.
type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	FullName  string    `json:"full_name"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// authResponse is returned by register (and later login).
type authResponse struct {
	User        userResponse `json:"user"`
	AccessToken string       `json:"access_token"`
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
		User:        toUserResponse(r.User),
		AccessToken: r.AccessToken,
	}
}
