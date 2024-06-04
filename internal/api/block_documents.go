package api

import (
	"context"

	"github.com/google/uuid"
)

type BlockDocumentClient interface {
	Get(ctx context.Context, id uuid.UUID) (*BlockDocument, error)
	Create(ctx context.Context, payload BlockDocumentCreate) (*BlockDocument, error)
	Update(ctx context.Context, id uuid.UUID, payload BlockDocumentUpdate) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetACL(ctx context.Context, id uuid.UUID) (*BlockDocumentAccess, error)
	UpdateACL(ctx context.Context, id uuid.UUID, payload BlockDocumentAccessReplace) error
}

type BlockDocument struct {
	BaseModel
	Name string `json:"name"`
	Data string `json:"data"`

	BlockSchemaID uuid.UUID    `json:"block_schema_id"`
	BlockSchema   *BlockSchema `json:"block_schema"`

	BlockTypeID   uuid.UUID `json:"block_type_id"`
	BlockTypeName *string   `json:"block_type_name"`
	BlockType     BlockType `json:"block_type"`

	BlockDocumentReferences string `json:"block_document_references"`
}

type BlockDocumentCreate struct {
	Name          string    `json:"name"`
	Data          string    `json:"data"`
	BlockSchemaID uuid.UUID `json:"block_schema_id"`
	BlockTypeID   uuid.UUID `json:"block_type_id"`
}

type BlockDocumentUpdate struct {
	BlockSchemaID     *uuid.UUID `json:"block_schema_id"`
	Data              string     `json:"data"`
	MergeExistingData bool       `json:"merge_existing_data"`
}

// BlockDocumentAccessReplace is the "update" request payload
// to modify a block document's current access control levels,
// meaning it contains the list of actors/teams + their respective access
// to a given block document.
type BlockDocumentAccessReplace struct {
	ManageActorIDs []AccessActorID `json:"manage_actor_ids"`
	ViewActorIDs   []AccessActorID `json:"view_actor_ids"`
	ManageTeamIDs  []uuid.UUID     `json:"manage_team_ids"`
	ViewTeamIDs    []uuid.UUID     `json:"view_team_ids"`
}

// BlockDocumentAccess is the API object representing a
// block document's current access control levels
// by actor (user/team/service account) and role (manage/view).
type BlockDocumentAccess struct {
	ManageActors []ObjectActorAccess `json:"manage_actors"`
	ViewActors   []ObjectActorAccess `json:"view_actors"`
}
