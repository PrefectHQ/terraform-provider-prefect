package api

import (
	"context"

	"github.com/google/uuid"
)

type WorkspaceAccessClient interface {
	Upsert(ctx context.Context, accessorType string, accessorID uuid.UUID, roleID uuid.UUID) (*WorkspaceAccess, error)
	Get(ctx context.Context, accessorType string, accessID uuid.UUID) (*WorkspaceAccess, error)
	Delete(ctx context.Context, accessorType string, accessID uuid.UUID) error
}

// WorkspaceAccess is a representation of a workspace access.
// This is used for multiple accessor types (user, service account, team),
// which dictates the presence of the specific accessor's ID.
type WorkspaceAccess struct {
	BaseModel
	WorkspaceID     uuid.UUID `json:"workspace_id"`
	WorkspaceRoleID uuid.UUID `json:"workspace_role_id"`

	ActorID *uuid.UUID `json:"actor_id"`
	BotID   *uuid.UUID `json:"bot_id"`
	UserID  *uuid.UUID `json:"user_id"`
}

// WorkspaceAccessUpsert defines the payload
// when upserting a workspace access request.
type WorkspaceAccessUpsert struct {
	WorkspaceRoleID uuid.UUID `json:"workspace_role_id"`

	// Only one of the follow IDs should be set on each call
	// depending on the resource's AccessorType
	// NOTE: omitempty normally excludes any zero value,
	// for primitives, but complex types like structs
	// and uuid.UUID require a pointer type to be omitted.
	UserID *uuid.UUID `json:"user_id,omitempty"`
	BotID  *uuid.UUID `json:"bot_id,omitempty"`
}
