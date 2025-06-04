package api

import (
	"context"

	"github.com/google/uuid"
)

// BlockSchemaClient is a client for working with block schemas.
type BlockSchemaClient interface {
	List(ctx context.Context, blockTypeIDs []uuid.UUID) ([]*BlockSchema, error)
	Create(ctx context.Context, payload *BlockSchemaCreate) (*BlockSchema, error)
	Read(ctx context.Context, id uuid.UUID) (*BlockSchema, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// BlockSchema is a representation of a block schema.
type BlockSchema struct {
	BaseModel

	BlockType    BlockType   `json:"block_type"`
	Checksum     string      `json:"checksum"`
	BlockTypeID  uuid.UUID   `json:"block_type_id"`
	Capabilities []string    `json:"capabilities"`
	Version      string      `json:"version"`
	Fields       interface{} `json:"fields"`
}

// BlockSchemaCreate is the create request payload.
type BlockSchemaCreate struct {
	BlockTypeID  uuid.UUID   `json:"block_type_id"`
	Capabilities []string    `json:"capabilities"`
	Version      string      `json:"version"`
	Fields       interface{} `json:"fields"`
}

// BlockSchemaFilter defines the search filter payload
// when searching for block schemas by slug.
type BlockSchemaFilter struct {
	// BlockSchemas
	BlockSchemas struct {
		BlockTypeID struct {
			Any []uuid.UUID `json:"any_"`
		} `json:"block_type_id"`
		BlockCapabilities struct {
			All []string `json:"all_"`
		} `json:"block_capabilities"`
	} `json:"block_schemas"`
}
