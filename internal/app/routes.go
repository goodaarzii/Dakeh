package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/repository/postgres"
	"github.com/goodaarzii/dakeh/internal/security"
	userhttp "github.com/goodaarzii/dakeh/internal/transport/http/users"
	userusecase "github.com/goodaarzii/dakeh/internal/usecase/user"
)

// Mount is the composition root: it builds every object and connects them.
// Flow for each feature: repository -> usecase -> handler -> routes.
func (a *Application) Mount() http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	api := r.Group("/api/v1")

	queries := sqlc.New(a.DB)

	// users
	userRepo := postgres.NewUserRepo(queries)
	userSvc := userusecase.NewService(userRepo, security.NewBcryptHasher())
	userhttp.NewHandler(userSvc).UsersRoutes(api)

	// products, cart, orders... follow the same four lines.

	return r
}
