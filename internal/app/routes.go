package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/middleware"
	"github.com/goodaarzii/dakeh/internal/repository/postgres"
	"github.com/goodaarzii/dakeh/internal/security"
	addresshttp "github.com/goodaarzii/dakeh/internal/transport/http/address"
	categoryhttp "github.com/goodaarzii/dakeh/internal/transport/http/category"
	producthttp "github.com/goodaarzii/dakeh/internal/transport/http/product"
	userhttp "github.com/goodaarzii/dakeh/internal/transport/http/users"
	addressusecase "github.com/goodaarzii/dakeh/internal/usecase/address"
	categoryusecase "github.com/goodaarzii/dakeh/internal/usecase/category"
	productusecase "github.com/goodaarzii/dakeh/internal/usecase/product"
	userusecase "github.com/goodaarzii/dakeh/internal/usecase/user"
)

// Mount is the composition root: it builds every object and connects them.
func (a *Application) Mount() http.Handler {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	api := r.Group("/api/v1")

	queries := sqlc.New(a.DB)

	// shared helpers
	hasher := security.NewBcryptHasher()
	tokens := security.NewJWTIssuer(a.Config.JWTSecret, a.Config.JWTTTL)

	// Three levels of access:
	//   api       -> public (no token)
	//   protected -> any logged-in user (Auth: 401 if no valid token)
	//   admin     -> logged-in AND role admin (RequireRole: 403 otherwise), URL prefix /admin
	protected := api.Group("", middleware.Auth(tokens))
	admin := protected.Group("/admin", middleware.RequireRole(domain.RoleAdmin))

	// users (register and login are public)
	userRepo := postgres.NewUserRepo(queries)
	userSvc := userusecase.NewService(userRepo, hasher, tokens)
	userhttp.NewHandler(userSvc).RegisterUserRoutes(api)

	// addresses (any logged-in user, their own data only)
	addressRepo := postgres.NewAddressRepo(queries)
	addressSvc := addressusecase.NewService(addressRepo)
	addresshttp.NewHandler(addressSvc).RegisterAddressRoutes(protected)

	// products (admin only for now)
	productRepo := postgres.NewProductRepo(queries)
	productSvc := productusecase.NewService(productRepo)
	producthttp.NewHandler(productSvc).RegisterProductRoutes(admin)

	// categories (admin only for now)
	categoryRepo := postgres.NewCategoryRepo(queries)
	categorySvc := categoryusecase.NewService(categoryRepo)
	categoryhttp.NewHandler(categorySvc).RegisterCategoryRoutes(admin)

	return r
}
