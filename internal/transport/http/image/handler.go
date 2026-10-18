package image

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
	"github.com/goodaarzii/dakeh/internal/transport/http/response"
	imageusecase "github.com/goodaarzii/dakeh/internal/usecase/image"
)

type Service interface {
	Upload(ctx context.Context, in imageusecase.UploadInput) (domain.ProductImage, error)
	ListByProduct(ctx context.Context, productID uuid.UUID) ([]domain.ProductImage, error)
	Update(ctx context.Context, id uuid.UUID, altText string, sortOrder int) (domain.ProductImage, error)
	SetPrimary(ctx context.Context, imageID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterImageRoutes RegisterImageRoutes: ADMIN routes. Pass the admin group (the /admin prefix and the role check come from it).
//
//	POST   /api/v1/admin/product/:id/image     multipart upload
//	GET    /api/v1/admin/product/:id/image
//	PUT    /api/v1/admin/image/:id
//	PUT    /api/v1/admin/image/:id/primary
//	DELETE /api/v1/admin/image/:id
func (h *Handler) RegisterImageRoutes(r *gin.RouterGroup) {
	r.POST("/product/:id/images", h.Upload)
	r.GET("/product/:id/images", h.ListByProduct)

	g := r.Group("/image")
	g.PUT("/:id", h.Update)
	g.PUT("/:id/primary", h.SetPrimary)
	g.DELETE("/:id", h.Delete)
}

func (h *Handler) Upload(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product id")
		return
	}

	// Stop reading after the maximum file size (+1 MB for the text fields),
	// so nobody can fill the server's memory with a giant upload.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, domain.MaxImageBytes+(1<<20))

	// The text fields of the form: alt_text, variant_id, sort_order.
	var form uploadImageForm
	if err := c.ShouldBind(&form); err != nil {
		writeUploadError(c, err)
		return
	}

	// The file field is called "file".
	fileHeader, err := c.FormFile("file")
	if err != nil {
		writeUploadError(c, err)
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_file", "cannot read the uploaded file")
		return
	}
	defer file.Close()

	img, err := h.svc.Upload(c.Request.Context(), imageusecase.UploadInput{
		ProductID: productID,
		VariantID: form.variantID(),
		AltText:   form.AltText,
		SortOrder: form.SortOrder,
		File:      file,
		Size:      fileHeader.Size,
	})
	if err != nil {
		response.WriteError(c, err)
		return
	}

	response.JSON(c, http.StatusCreated, toImageResponse(img))
}

func (h *Handler) ListByProduct(c *gin.Context) {
	productID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid product id")
		return
	}

	images, err := h.svc.ListByProduct(c.Request.Context(), productID)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toImageResponses(images))
}

func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid image id")
		return
	}

	var req updateImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {

		return
	}

	img, err := h.svc.Update(c.Request.Context(), id, req.AltText, req.SortOrder)
	if err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, toImageResponse(img))
}

func (h *Handler) SetPrimary(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid image id")
		return
	}

	if err := h.svc.SetPrimary(c.Request.Context(), id); err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, response.NewStatusResponse(1))
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_id", "invalid image id")
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.WriteError(c, err)
		return
	}
	response.JSON(c, http.StatusOK, response.NewStatusResponse(1))

}

// writeUploadError turns "the body was bigger than the limit" into 413, everything else into 400.
func writeUploadError(c *gin.Context, err error) {
	var tooBig *http.MaxBytesError
	if errors.As(err, &tooBig) {
		response.Error(c, http.StatusRequestEntityTooLarge, "image_too_large", "the upload is too large")
		return
	}

}
