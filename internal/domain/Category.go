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
