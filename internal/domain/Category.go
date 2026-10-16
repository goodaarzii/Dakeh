package domain

import (
	"time"

	"github.com/google/uuid"
)

// Category is a node in the category tree. ParentID is nil for top-level categories.
type Category struct {
	ID          uuid.UUID
	ParentID    *uuid.UUID
	Name        string
	Slug        string
	Description string
	ImageURL    string
	SortOrder   int
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CategoryNode is a category together with its children.
// The usecase builds the tree from the flat list that the repository returns.
type CategoryNode struct {
	Category
	Children []CategoryNode
}

// BuildCategoryTree turns the flat list from the repository into a tree.
// The input order is kept (the query already sorts by sort_order, then name).
//
// Only categories reachable from a top-level category (ParentID == nil) are included.
// So a category whose parent is missing from the list (for example an INACTIVE parent)
// disappears together with its whole subtree: a hidden parent hides its children.
// Every Children slice is non-nil, so JSON shows [] and never null.
func BuildCategoryTree(all []Category) []CategoryNode {
	childrenOf := make(map[uuid.UUID][]Category)
	var roots []Category
	for _, c := range all {
		if c.ParentID == nil {
			roots = append(roots, c)
			continue
		}
		childrenOf[*c.ParentID] = append(childrenOf[*c.ParentID], c)
	}

	var build func(cats []Category) []CategoryNode
	build = func(cats []Category) []CategoryNode {
		nodes := make([]CategoryNode, len(cats))
		for i, c := range cats {
			nodes[i] = CategoryNode{Category: c, Children: build(childrenOf[c.ID])}
		}
		return nodes
	}
	return build(roots)
}

// CategoryAndDescendantIDs returns the ID of the category with this slug plus the IDs of all its
// children, grandchildren, and so on. Used for the product list filter: asking for "shoes"
// must also return the products that sit in "running shoes". found is false for an unknown slug.
func CategoryAndDescendantIDs(all []Category, slug string) (ids []uuid.UUID, found bool) {
	childrenOf := make(map[uuid.UUID][]uuid.UUID)
	var start *Category
	for i := range all {
		if all[i].Slug == slug {
			start = &all[i]
		}
		if all[i].ParentID != nil {
			childrenOf[*all[i].ParentID] = append(childrenOf[*all[i].ParentID], all[i].ID)
		}
	}
	if start == nil {
		return nil, false
	}

	// Breadth-first walk. "seen" protects against a loop in bad data (A is the parent of B and B of A),
	// so a public endpoint can never be made to spin forever.
	seen := map[uuid.UUID]bool{start.ID: true}
	ids = []uuid.UUID{start.ID}
	for i := 0; i < len(ids); i++ {
		for _, child := range childrenOf[ids[i]] {
			if !seen[child] {
				seen[child] = true
				ids = append(ids, child)
			}
		}
	}
	return ids, true
}
