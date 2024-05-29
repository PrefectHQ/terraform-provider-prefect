package api

import (
	"context"
)

// BlockTypeClient is a client for working with block types.
type BlockTypeClient interface {
	GetBySlug(ctx context.Context, slug string) (*BlockType, error)
}

// BlockType is a representation of a block type.
type BlockType struct {
	BaseModel
	Slug string `json:"slug"`
}
