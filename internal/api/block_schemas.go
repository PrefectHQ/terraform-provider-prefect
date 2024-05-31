package api

import (
	"context"

	"github.com/google/uuid"
)

// BlockSchemaClient is a client for working with block schemas.
type BlockSchemaClient interface {
	List(ctx context.Context, blockTypeIDs []uuid.UUIDs) ([]*BlockSchema, error)
}

// BlockSchema is a representation of a block schema.
type BlockSchema struct {
	BaseModel
	BlockType

	Checksum     string      `json:"checksum"`
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
			Any []uuid.UUIDs `json:"any_"`
		} `json:"block_type_id"`
		BlockCapabilities struct {
			All []string `json:"all_"`
		} `json:"block_capabilities"`
	} `json:"block_schemas"`
}
