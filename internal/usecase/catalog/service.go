package catalog

import (
	"context"
	"strings"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// Repository is what this usecase needs from storage. *postgres.CatalogRepo satisfies it.
type Repository interface {
	ListActiveCategories(ctx context.Context) ([]domain.Category, error)
	GetTags(ctx context.Context) ([]domain.Tag, error)
	ListPublicProducts(ctx context.Context, q domain.ProductQuery) ([]domain.ProductCard, error)
	GetActiveProductBySlug(ctx context.Context, slug string) (domain.ProductDetail, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// CategoryTree returns the active categories as a tree (menus).
func (s *Service) CategoryTree(ctx context.Context) ([]domain.CategoryNode, error) {
	all, err := s.repo.ListActiveCategories(ctx)
	if err != nil {
		return nil, err
	}
	return domain.BuildCategoryTree(all), nil
}

func (s *Service) GetTags(ctx context.Context) ([]domain.Tag, error) {
	return s.repo.GetTags(ctx)
}

// GetActiveProductBySlug returns the whole product page. The repo does the work.
// An unknown, draft, or variant-less product comes back as domain.ErrNotFound (-> 404).
func (s *Service) GetActiveProductBySlug(ctx context.Context, slug string) (domain.ProductDetail, error) {
	return s.repo.GetActiveProductBySlug(ctx, slug)
}

// ListProducts turns the shop filter (slugs, page numbers) into a repository query (IDs, limit, offset).
// hasMore tells the frontend whether a next page exists.
func (s *Service) ListProducts(ctx context.Context, f domain.ProductFilter) ([]domain.ProductCard, bool, error) {
	// Do not trust the caller: clamp the paging and the sort even though the handler already did.
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PageSize < 1 {
		f.PageSize = 20
	}
	if f.PageSize > 50 {
		f.PageSize = 50
	}
	switch f.Sort {
	case domain.SortNewest, domain.SortPriceAsc, domain.SortPriceDesc:
	default:
		f.Sort = domain.SortNewest
	}

	q := domain.ProductQuery{
		TagSlug: f.TagSlug,
		Search:  escapeLike(f.Search),
		Sort:    f.Sort,
		// Ask for ONE extra row. If it comes back, a next page exists.
		Limit:  f.PageSize + 1,
		Offset: (f.Page - 1) * f.PageSize,
	}

	// Category filter: slug -> the category and all its children (as IDs).
	if f.CategorySlug != "" {
		all, err := s.repo.ListActiveCategories(ctx)
		if err != nil {
			return nil, false, err
		}
		ids, found := domain.CategoryAndDescendantIDs(all, f.CategorySlug)
		if !found {
			return []domain.ProductCard{}, false, nil // unknown category: an empty page, not an error
		}
		q.CategoryIDs = ids
	}

	cards, err := s.repo.ListPublicProducts(ctx, q)
	if err != nil {
		return nil, false, err
	}

	hasMore := len(cards) > f.PageSize
	if hasMore {
		cards = cards[:f.PageSize] // drop the extra row
	}
	return cards, hasMore, nil
}

// escapeLike makes "%" and "_" typed by a customer count as normal characters in the ILIKE search.
// Backslash is Postgres' default escape character for LIKE.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}
