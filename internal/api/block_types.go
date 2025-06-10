package api

import (
	"context"

	"github.com/google/uuid"
)

// BlockTypeClient is a client for working with block types.
type BlockTypeClient interface {
	GetBySlug(ctx context.Context, slug string) (*BlockType, error)
	Create(ctx context.Context, payload *BlockTypeCreate) (*BlockType, error)
	Update(ctx context.Context, id uuid.UUID, payload *BlockTypeUpdate) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// BlockType is a representation of a block type.
type BlockType struct {
	BaseModel

	Name             string `json:"name"`
	Slug             string `json:"slug"`
	LogoURL          string `json:"logo_url"`
	DocumentationURL string `json:"documentation_url"`
	Description      string `json:"description"`
	CodeExample      string `json:"code_example"`

	IsProtected bool `json:"is_protected"`
}

// BlockTypeCreate is the create request payload.
type BlockTypeCreate struct {
	Name             string `json:"name"`
	Slug             string `json:"slug"`
	LogoURL          string `json:"logo_url"`
	DocumentationURL string `json:"documentation_url"`
	Description      string `json:"description"`
	CodeExample      string `json:"code_example"`
}

// BlockTypeUpdate is the update request payload.
type BlockTypeUpdate struct {
	LogoURL          string `json:"logo_url"`
	DocumentationURL string `json:"documentation_url"`
	Description      string `json:"description"`
	CodeExample      string `json:"code_example"`
}
