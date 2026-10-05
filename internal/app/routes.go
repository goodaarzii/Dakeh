package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goodaarzii/dakeh/db/sqlc"
	"github.com/goodaarzii/dakeh/internal/middleware"
	"github.com/goodaarzii/dakeh/internal/repository/postgres"
	"github.com/goodaarzii/dakeh/internal/security"
	"github.com/goodaarzii/dakeh/internal/transport/http/address"
	category2 "github.com/goodaarzii/dakeh/internal/transport/http/category"
	product2 "github.com/goodaarzii/dakeh/internal/transport/http/product"
	userhttp "github.com/goodaarzii/dakeh/internal/transport/http/users"
	address2 "github.com/goodaarzii/dakeh/internal/usecase/address"
	"github.com/goodaarzii/dakeh/internal/usecase/category"
	"github.com/goodaarzii/dakeh/internal/usecase/product"
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

	//middleware auth
	protected := api.Group("", middleware.Auth(tokens))

	// users
	userRepo := postgres.NewUserRepo(queries)
	userSvc := userusecase.NewService(userRepo, hasher, tokens)
	userhttp.NewHandler(userSvc).RegisterUserRoutes(api)

	//address
	addrRepo := postgres.NewAddressRepo(queries)
	addrUsaCase := address2.NewService(addrRepo)
	address.NewHandler(addrUsaCase).RegisterAddressRoutes(protected)

	//product
	productRepo := postgres.NewProductRepo(queries)
	productUseCase := product.NewService(productRepo)
	product2.NewHandler(productUseCase).RegisterProductRoutes(protected)

	//category
	categoryRepo := postgres.NewCategoryRepo(queries)
	categoryUseCase := category.NewService(categoryRepo)
	category2.NewHandler(categoryUseCase).RegisterAddressRoutes(protected)

	return r
}
