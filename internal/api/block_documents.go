package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
	Name *string `json:"name"` // names are optional for anonymous blocks
	Data string  `json:"data"`

	BlockSchemaID uuid.UUID `json:"block_schema_id"`
	// BlockSchema   *BlockSchema `json:"block_schema"`

	BlockTypeID   uuid.UUID `json:"block_type_id"`
	BlockTypeName *string   `json:"block_type_name"`
	BlockType     BlockType `json:"block_type"`

	BlockDocumentReferences string `json:"block_document_references"`

	IsAnonymous bool `json:"is_anonymous"`
}

type BlockDocumentCreate struct {
	Name          *string   `json:"name"` // names are optional for anonymous blocks
	Data          string    `json:"data"`
	BlockSchemaID uuid.UUID `json:"block_schema_id"`
	BlockTypeID   uuid.UUID `json:"block_type_id"`
	IsAnonymous   bool      `json:"is_anonymous"`
}

type BlockDocumentUpdate struct {
	BlockSchemaID     *uuid.UUID `json:"block_schema_id"`
	Data              string     `json:"data"`
	MergeExistingData bool       `json:"merge_existing_data"`
}

// AccessControlList is a custom type that represents
// an API response where the value can be:
// []uuid.UUID - a list of IDs
// ["*"] - a wildcard, meaning "all".
//
// nolint:musttag // we have custom marshal/unmarshal logic for this type
type AccessControlList struct {
	IDs []uuid.UUID
	All bool
}

// Custom JSON marshaling for AccessControlList
// so we can return []uuid.UUID or ["*"] back to the API.
func (acl AccessControlList) MarshalJSON() ([]byte, error) {
	if acl.All {
		data, err := json.Marshal([]string{"*"})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal wildcard ACL: %w", err)
		}

		return data, nil
	}

	data, err := json.Marshal(acl.IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ID slice ACL: %w", err)
	}

	return data, nil
}

// Custom JSON unmarshaling for AccessControlList
// so we can accept []uuid.UUID or ["*"] from the API
// in a structured format.
func (acl *AccessControlList) UnmarshalJSON(data []byte) error {
	var ids []uuid.UUID
	if err := json.Unmarshal(data, &ids); err == nil {
		acl.IDs = ids
		acl.All = false

		return nil
	}

	var all []string
	if err := json.Unmarshal(data, &all); err == nil && len(all) == 1 && all[0] == "*" {
		acl.All = true
		acl.IDs = nil

		return nil
	}

	return errors.New("invalid AccessControlList format")
}

// BlockDocumentAccessReplace is the "update" request payload
// to modify a block document's current access control levels,
// meaning it contains the list of actors/teams + their respective access
// to a given block document.
type BlockDocumentAccessReplace struct {
	ManageActorIDs AccessControlList `json:"manage_actor_ids"`
	ViewActorIDs   AccessControlList `json:"view_actor_ids"`
	ManageTeamIDs  []uuid.UUID       `json:"manage_team_ids"`
	ViewTeamIDs    []uuid.UUID       `json:"view_team_ids"`
}

// BlockActorType represents an enum of type values
// used in the Block Document Access API.
type BlockActorType string

const (
	BlockUser           BlockActorType = "user"
	BlockServiceAccount BlockActorType = "service_account"
	BlockTeam           BlockActorType = "team"
	BlockAll            BlockActorType = "*"
)

type ObjectActorAccess struct {
	ID    AccessControlList `json:"id"`
	Name  string            `json:"name"`
	Email *string           `json:"email"`
	Type  BlockActorType    `json:"type"`
}

// BlockDocumentAccess is the API object representing a
// block document's current access control levels
// by actor (user/team/service account) and role (manage/view).
type BlockDocumentAccess struct {
	ManageActors []ObjectActorAccess `json:"manage_actors"`
	ViewActors   []ObjectActorAccess `json:"view_actors"`
}
