package api

import (
	"context"

	"github.com/google/uuid"
)

type BlockDocumentClient interface {
	Get(ctx context.Context, id uuid.UUID) (*BlockDocument, error)
	GetByName(ctx context.Context, typeSlug, name string) (*BlockDocument, error)
	Create(ctx context.Context, payload BlockDocumentCreate) (*BlockDocument, error)
	Update(ctx context.Context, id uuid.UUID, payload BlockDocumentUpdate) error
	Delete(ctx context.Context, id uuid.UUID) error

	GetAccess(ctx context.Context, id uuid.UUID) (*BlockDocumentAccess, error)
	UpsertAccess(ctx context.Context, id uuid.UUID, payload BlockDocumentAccessUpsert) error
}

type BlockDocument struct {
	BaseModel
	Name string                 `json:"name"`
	Data map[string]interface{} `json:"data"`

	BlockSchemaID uuid.UUID    `json:"block_schema_id"`
	BlockSchema   *BlockSchema `json:"block_schema"`

	BlockTypeID   uuid.UUID `json:"block_type_id"`
	BlockTypeName *string   `json:"block_type_name"`
	BlockType     BlockType `json:"block_type"`
}

type BlockDocumentCreate struct {
	Name          string                 `json:"name"`
	Data          map[string]interface{} `json:"data"`
	BlockSchemaID uuid.UUID              `json:"block_schema_id"`
	BlockTypeID   uuid.UUID              `json:"block_type_id"`
}

type BlockDocumentUpdate struct {
	BlockSchemaID     uuid.UUID              `json:"block_schema_id"`
	Data              map[string]interface{} `json:"data"`
	MergeExistingData bool                   `json:"merge_existing_data"`
}

// BlockDocumentAccessUpsert is the create/update request payload
// to modify a block document's current access control levels,
// meaning it contains the list of actors/teams + their respective access
// to a given block document.
type BlockDocumentAccessUpsert struct {
	AccessControl struct {
		ManageActorIDs []string `json:"manage_actor_ids"`
		ViewActorIDs   []string `json:"view_actor_ids"`
		ManageTeamIDs  []string `json:"manage_team_ids"`
		ViewTeamIDs    []string `json:"view_team_ids"`
	} `json:"access_control"`
}

// BlockDocumentAccess is the API object representing a
// block document's current access control levels
// by actor (user/team/service account) and role (manage/view).
type BlockDocumentAccess struct {
	ManageActors []ObjectActorAccess `json:"manage_actors"`
	ViewActors   []ObjectActorAccess `json:"view_actors"`
}
