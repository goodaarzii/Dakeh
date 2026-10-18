package cart

import (
	"context"
	"fmt"
	"sort"

	"github.com/google/uuid"

	"github.com/goodaarzii/dakeh/internal/domain"
)

// ItemStore is where the cart lives (Redis): only variant IDs and quantities.
type ItemStore interface {
	Get(ctx context.Context, userID uuid.UUID) ([]domain.CartItem, error)
	SetItems(ctx context.Context, userID uuid.UUID, items []domain.CartItem) error
	RemoveItem(ctx context.Context, userID, variantID uuid.UUID) error
	Clear(ctx context.Context, userID uuid.UUID) error
}

// Catalog gives the live data (price, name, stock) from Postgres. *postgres.CatalogRepo satisfies it.
type Catalog interface {
	ListSellableVariants(ctx context.Context, variantIDs []uuid.UUID) ([]domain.SellableVariant, error)
}

type Service struct {
	store   ItemStore
	catalog Catalog
}

func NewService(store ItemStore, catalog Catalog) *Service {
	return &Service{store: store, catalog: catalog}
}

// Get returns the cart with LIVE prices and stock.
func (s *Service) Get(ctx context.Context, userID uuid.UUID) (domain.Cart, error) {
	items, err := s.store.Get(ctx, userID)
	if err != nil {
		return domain.Cart{}, err
	}
	return s.build(ctx, items)
}

// SetQuantity puts a variant in the cart with exactly this quantity (adds it, or changes it).
func (s *Service) SetQuantity(ctx context.Context, userID, variantID uuid.UUID, quantity int) (domain.Cart, error) {
	if quantity < 1 || quantity > domain.MaxQuantityPerLine {
		return domain.Cart{}, fmt.Errorf("%w: quantity must be between 1 and %d", domain.ErrInvalidInput, domain.MaxQuantityPerLine)
	}

	sellable, err := s.catalog.ListSellableVariants(ctx, []uuid.UUID{variantID})
	if err != nil {
		return domain.Cart{}, err
	}
	if len(sellable) == 0 {
		return domain.Cart{}, domain.ErrVariantUnavailable
	}
	if quantity > sellable[0].Available {
		return domain.Cart{}, domain.ErrInsufficientStock
	}

	items, err := s.store.Get(ctx, userID)
	if err != nil {
		return domain.Cart{}, err
	}
	if len(items) >= domain.MaxCartLines && !contains(items, variantID) {
		return domain.Cart{}, domain.ErrCartFull
	}

	if err := s.store.SetItems(ctx, userID, []domain.CartItem{{VariantID: variantID, Quantity: quantity}}); err != nil {
		return domain.Cart{}, err
	}
	return s.Get(ctx, userID)
}

func (s *Service) RemoveItem(ctx context.Context, userID, variantID uuid.UUID) (domain.Cart, error) {
	if err := s.store.RemoveItem(ctx, userID, variantID); err != nil {
		return domain.Cart{}, err
	}
	return s.Get(ctx, userID)
}

func (s *Service) Clear(ctx context.Context, userID uuid.UUID) error {
	return s.store.Clear(ctx, userID)
}

// Merge adds the items of a GUEST cart (kept in the browser until login) to the user's cart.
// Rules: quantities of the same variant are ADDED, then capped by the per-line limit and by the stock.
// Items that are no longer sellable are skipped silently: a login must never fail because of an old cart.
// Calling it twice with the same items adds them twice, so the frontend must clear its guest cart afterwards.
func (s *Service) Merge(ctx context.Context, userID uuid.UUID, guest []domain.CartItem) (domain.Cart, error) {
	// 1. one entry per variant
	want := make(map[uuid.UUID]int)
	for _, it := range guest {
		if it.Quantity > 0 {
			want[it.VariantID] += it.Quantity
		}
	}
	if len(want) == 0 {
		return s.Get(ctx, userID)
	}

	existing, err := s.store.Get(ctx, userID)
	if err != nil {
		return domain.Cart{}, err
	}
	have := make(map[uuid.UUID]int, len(existing))
	for _, it := range existing {
		have[it.VariantID] = it.Quantity
	}

	// 2. live data of every variant involved
	ids := make([]uuid.UUID, 0, len(want))
	for id := range want {
		ids = append(ids, id)
	}
	sellable, err := s.catalog.ListSellableVariants(ctx, ids)
	if err != nil {
		return domain.Cart{}, err
	}
	byID := make(map[uuid.UUID]domain.SellableVariant, len(sellable))
	for _, v := range sellable {
		byID[v.VariantID] = v
	}

	// 3. decide the new quantities
	lines := len(have)
	var toSet []domain.CartItem
	for variantID, add := range want {
		v, ok := byID[variantID]
		if !ok {
			continue // no longer sold
		}
		current, inCart := have[variantID]
		if !inCart && lines >= domain.MaxCartLines {
			continue // the cart is full
		}

		q := current + add
		if q > domain.MaxQuantityPerLine {
			q = domain.MaxQuantityPerLine
		}
		if q > v.Available {
			q = v.Available
		}
		if q <= current { // nothing to gain (also: never REDUCE a quantity the user already had)
			continue
		}

		if !inCart {
			lines++
		}
		toSet = append(toSet, domain.CartItem{VariantID: variantID, Quantity: q})
	}

	if err := s.store.SetItems(ctx, userID, toSet); err != nil {
		return domain.Cart{}, err
	}
	return s.Get(ctx, userID)
}

// build joins the stored items with the live data from Postgres.
func (s *Service) build(ctx context.Context, items []domain.CartItem) (domain.Cart, error) {
	if len(items) == 0 {
		return domain.Cart{Lines: []domain.CartLine{}}, nil
	}

	ids := make([]uuid.UUID, len(items))
	for i, it := range items {
		ids[i] = it.VariantID
	}
	sellable, err := s.catalog.ListSellableVariants(ctx, ids) // ONE query for the whole cart
	if err != nil {
		return domain.Cart{}, err
	}
	byID := make(map[uuid.UUID]domain.SellableVariant, len(sellable))
	for _, v := range sellable {
		byID[v.VariantID] = v
	}

	cart := domain.Cart{Lines: make([]domain.CartLine, len(items))}
	allGood := true
	for i, it := range items {
		line := domain.CartLine{VariantID: it.VariantID, Quantity: it.Quantity}
		if v, ok := byID[it.VariantID]; ok {
			line.Sellable = true
			line.Variant = v
			line.InStock = v.Available >= it.Quantity
			line.LineTotal = v.UnitPrice * int64(it.Quantity)
			cart.Subtotal += line.LineTotal
		}
		if !line.Sellable || !line.InStock {
			allGood = false
		}
		cart.ItemCount += it.Quantity
		cart.Lines[i] = line
	}
	cart.CanCheckout = allGood

	// A Redis hash has no order, so sort: sellable lines first, then by product and variant name.
	sort.Slice(cart.Lines, func(a, b int) bool {
		la, lb := cart.Lines[a], cart.Lines[b]
		if la.Sellable != lb.Sellable {
			return la.Sellable
		}
		if la.Variant.ProductName != lb.Variant.ProductName {
			return la.Variant.ProductName < lb.Variant.ProductName
		}
		if la.Variant.VariantName != lb.Variant.VariantName {
			return la.Variant.VariantName < lb.Variant.VariantName
		}
		return la.VariantID.String() < lb.VariantID.String()
	})
	return cart, nil
}

func contains(items []domain.CartItem, variantID uuid.UUID) bool {
	for _, it := range items {
		if it.VariantID == variantID {
			return true
		}
	}
	return false
}
