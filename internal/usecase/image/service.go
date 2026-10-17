package image

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// Repository is what this usecase needs from the product_images table (postgres.ImageRepo).
type Repository interface {
	Create(ctx context.Context, img domain.ProductImage) (domain.ProductImage, error)
	CountByProduct(ctx context.Context, productID uuid.UUID) (int, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.ProductImage, error)
	ListByProduct(ctx context.Context, productID uuid.UUID) ([]domain.ProductImage, error)
	Update(ctx context.Context, id uuid.UUID, altText string, sortOrder int) (domain.ProductImage, error)
	SetPrimary(ctx context.Context, productID, imageID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) (domain.ProductImage, error)
}

// Storage is where the files live (*storage.MinioStorage satisfies it).
type Storage interface {
	Put(ctx context.Context, key, contentType string, r io.Reader, size int64) error
	Delete(ctx context.Context, key string) error
	PublicURL(key string) string
}

type Service struct {
	repo    Repository
	storage Storage
}

func NewService(repo Repository, storage Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

// UploadInput: File is a ReadSeeker because we first LOOK at the first bytes, then read from the start.
type UploadInput struct {
	ProductID uuid.UUID
	VariantID *uuid.UUID // optional
	AltText   string
	SortOrder int
	File      io.ReadSeeker
	Size      int64
}

// Upload validates the file, stores it, and saves its row.
func (s *Service) Upload(ctx context.Context, in UploadInput) (domain.ProductImage, error) {
	// 1. cheap checks first
	if in.Size <= 0 {
		return domain.ProductImage{}, domain.ErrUnsupportedImage // empty file
	}
	if in.Size > domain.MaxImageBytes {
		return domain.ProductImage{}, domain.ErrImageTooLarge
	}

	// 2. what is this file REALLY? Look at its first bytes. Never trust the file name or the client's header.
	head := make([]byte, 512)
	n, err := io.ReadFull(in.File, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return domain.ProductImage{}, fmt.Errorf("read upload: %w", err)
	}
	contentType := http.DetectContentType(head[:n])
	ext, ok := domain.ImageExtension(contentType)
	if !ok {
		return domain.ProductImage{}, domain.ErrUnsupportedImage
	}
	if _, err := in.File.Seek(0, io.SeekStart); err != nil {
		return domain.ProductImage{}, fmt.Errorf("rewind upload: %w", err)
	}

	// 3. the product may not have too many image (checked BEFORE uploading)
	count, err := s.repo.CountByProduct(ctx, in.ProductID)
	if err != nil {
		return domain.ProductImage{}, err
	}
	if count >= domain.MaxImagesPerProduct {
		return domain.ProductImage{}, domain.ErrTooManyImages
	}

	// 4. the key is ours: never use the client's file name (it can contain "../" or anything else)
	key := fmt.Sprintf("products/%s/%s%s", in.ProductID, uuid.New(), ext)
	if err := s.storage.Put(ctx, key, contentType, in.File, in.Size); err != nil {
		return domain.ProductImage{}, err
	}

	// 5. save the row. The first image of a product becomes primary (the repo does that).
	img, err := s.repo.Create(ctx, domain.ProductImage{
		ProductID:  in.ProductID,
		VariantID:  in.VariantID,
		StorageKey: key,
		URL:        s.storage.PublicURL(key),
		AltText:    strings.TrimSpace(in.AltText),
		SortOrder:  in.SortOrder,
	})
	if err != nil {
		s.removeFile(key) // the row was not saved (unknown product, bad variant...), so do not leave the file behind
		return domain.ProductImage{}, err
	}
	return img, nil
}

func (s *Service) ListByProduct(ctx context.Context, productID uuid.UUID) ([]domain.ProductImage, error) {
	return s.repo.ListByProduct(ctx, productID)
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, altText string, sortOrder int) (domain.ProductImage, error) {
	return s.repo.Update(ctx, id, strings.TrimSpace(altText), sortOrder)
}

// SetPrimary only receives the image ID, so it first finds out which product the image belongs to.
func (s *Service) SetPrimary(ctx context.Context, imageID uuid.UUID) error {
	img, err := s.repo.GetByID(ctx, imageID)
	if err != nil {
		return err
	}
	return s.repo.SetPrimary(ctx, img.ProductID, imageID)
}

// Delete removes the row first (and promotes another primary if needed), then the file.
// If the file cannot be deleted, the row is already gone, so we only log it: an orphan file is harmless
// and can be cleaned up later.
func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	img, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if err := s.storage.Delete(ctx, img.StorageKey); err != nil {
		slog.Warn("image row deleted but its file was not", "key", img.StorageKey, "err", err)
	}
	return nil
}

// removeFile is a best-effort cleanup. It uses its own context: the request's context may already be cancelled.
func (s *Service) removeFile(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.storage.Delete(ctx, key); err != nil {
		slog.Warn("could not remove an uploaded file after a failure", "key", key, "err", err)
	}
}
