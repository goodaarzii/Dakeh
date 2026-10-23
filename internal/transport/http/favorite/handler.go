package favorite

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
	Add(ctx context.Context, userID, productID uuid.UUID) (domain.Favorite, error)
	Remove(ctx context.Context, userID, productID uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID) ([]domain.FavoriteProduct, error)
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterFavoriteRoutes registers user-owned favorites. Call it with the protected API group.
//
//	GET    /api/v1/favorites
//	PUT    /api/v1/favorites/:product_id
//	DELETE /api/v1/favorites/:product_id
func (h *Handler) RegisterFavoriteRoutes(r *gin.RouterGroup) {
	g := r.Group("/favorites")
	g.GET("", h.List)
	g.PUT("/:product_id", h.Add)
	g.DELETE("/:product_id", h.Remove)
}

func (h *Handler) List(c *gin.Context) {
	favorites, err := h.svc.List(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toFavoritesListResponse(favorites))
}

func (h *Handler) Add(c *gin.Context) {
	productID, err := parseProductID(c)
	if err != nil {
		return
	}

	userID := middleware.UserID(c)
	if userID == uuid.Nil {
		response.Error(
			c,
			http.StatusUnauthorized,
			"unauthorized",
			"authenticated user ID is missing",
		)
		return
	}

	favorite, err := h.svc.Add(
		c.Request.Context(),
		userID,
		productID,
	)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toFavoriteResponse(favorite))
}

func (h *Handler) Remove(c *gin.Context) {
	productID, err := parseProductID(c)
	if err != nil {
		return
	}

	if err := h.svc.Remove(c.Request.Context(), middleware.UserID(c), productID); err != nil {
		response.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func parseProductID(c *gin.Context) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Param("product_id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product id")
		return uuid.Nil, err
	}
	return id, nil
}
