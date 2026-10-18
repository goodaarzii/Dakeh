package variant

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, v domain.Variant, initialStock, lowStockThreshold int) (domain.VariantWithStock, error)
	Update(ctx context.Context, v domain.Variant) (domain.VariantWithStock, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.VariantWithStock, error)
	ListByProduct(ctx context.Context, productID uuid.UUID) ([]domain.VariantWithStock, error)
	Delete(ctx context.Context, id uuid.UUID) error

	GetStock(ctx context.Context, variantID uuid.UUID) (domain.Inventory, error)
	ListStock(ctx context.Context, variantIDs []uuid.UUID) ([]domain.Inventory, error)
	SetStock(ctx context.Context, variantID uuid.UUID, onHand, lowStockThreshold int) (domain.Inventory, error)
	AdjustStock(ctx context.Context, variantID uuid.UUID, delta int) (domain.Inventory, error)
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

func (h *Handler) RegisterVariantRoutes(r *gin.RouterGroup) {
	r.POST("/product/:id/variants", h.Create)
	r.GET("/product/:id/variants", h.ListByProduct)

	g := r.Group("/variant")
	g.GET("/:id", h.GetByID)
	g.PUT("/:id", h.Update)
	g.DELETE("/:id", h.Delete)
	g.GET("/:id/stock", h.GetStock) // moved here
	g.PUT("/:id/stock", h.SetStock)
	g.POST("/:id/stock/adjust", h.AdjustStock)

	r.POST("/stock/query", h.ListStock) // this one has no ID in the URL, so it stays
}

func (h *Handler) Create(c *gin.Context) {
	// 1. product ID from the URL: /product/:id/variants
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product id")
		return
	}

	// 2. JSON body -> request struct (validation runs here)
	var req createVariantRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		return
	}

	// 3. request -> domain, then attach the product from the URL
	variant := req.toDomain()
	variant.ProductID = productID

	// 4. call the usecase
	created, err := h.svc.Create(c.Request.Context(), variant, req.InitialStock, req.LowStockThreshold)
	if err != nil {
		response.WriteError(c, err) // use the same error helper your other handlers use
		return
	}

	// 5. domain -> response, with 201 Created
	response.JSON(c, http.StatusCreated, toVariantResponse(created))
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product id")
		return
	}
	var req variantRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		return
	}
	v := req.toDomain() // handles IsActive (defaults to true) and nil attributes
	v.ID = id
	//handle * at the end project
	update, err := h.svc.Update(c.Request.Context(), v)

	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// make response for this below
	response.JSON(c, http.StatusOK, toVariantResponse(update))

}

func (h *Handler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid variant id")
		return
	}

	v, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, toVariantResponse(v))
}

func (h *Handler) ListByProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product id")
		return
	}
	variants, err := h.svc.ListByProduct(c.Request.Context(), productID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	ids := make([]uuid.UUID, len(variants))
	for i, v := range variants {
		ids[i] = v.Variant.ID
	}

	stocks, err := h.svc.ListStock(c.Request.Context(), ids)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	byVariant := make(map[uuid.UUID]domain.Inventory, len(stocks))
	for _, st := range stocks {
		byVariant[st.VariantID] = st
	}

	responseVariant := make([]domain.VariantWithStock, len(variants))
	for i, v := range variants {
		responseVariant[i] = domain.VariantWithStock{Variant: v.Variant, Stock: byVariant[v.Variant.ID]}
	}

	response.JSON(c, http.StatusOK, toVariantResponses(responseVariant))
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product id")
		return
	}
	err = h.svc.Delete(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.JSON(c, http.StatusOK, response.NewStatusResponse(1))
}

func (h *Handler) SetStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid variant id")
		return
	}

	var req setStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		return
	}

	stock, err := h.svc.SetStock(c.Request.Context(), id, *req.QuantityOnHand, req.LowStockThreshold)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, toStockResponse(stock))
}

func (h *Handler) AdjustStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid variant id")
		return
	}

	var req adjustStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		return
	}
	stock, err := h.svc.AdjustStock(c.Request.Context(), id, req.Delta)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toStockResponse(stock))

}

func (h *Handler) GetStock(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product id")
		return
	}
	variant, err := h.svc.GetStock(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "internal_error", err.Error())
		return

	}
	response.JSON(c, http.StatusOK, toStockResponse(variant))
}

func (h *Handler) ListStock(c *gin.Context) {
	var req listStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		return
	}

	stocks, err := h.svc.ListStock(c.Request.Context(), req.VariantIDs)
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.JSON(c, http.StatusOK, toVariantStockResponses(stocks))
}
