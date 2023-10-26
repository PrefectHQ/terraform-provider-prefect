package api

import (
	"context"

	"github.com/google/uuid"
)

type WorkspaceAccessClient interface {
	UpsertServiceAccountAccess(ctx context.Context, payload WorkspaceAccessUpsert) (*WorkspaceServiceAccountAccess, error)
	UpsertUserAccess(ctx context.Context, payload WorkspaceAccessUpsert) (*WorkspaceUserAccess, error)

	GetServiceAccountAccess(ctx context.Context, accessID uuid.UUID) (*WorkspaceServiceAccountAccess, error)
	GetUserAccess(ctx context.Context, accessID uuid.UUID) (*WorkspaceUserAccess, error)

	DeleteServiceAccountAccess(ctx context.Context, accessID uuid.UUID) error
	DeleteUserAccess(ctx context.Context, accessID uuid.UUID) error
}

// WorkspaceAccessBaseModel sets the shared attributes
// for different workspace accessor responses, such as
// users, service accounts, and teams.
type WorkspaceAccessBaseModel struct {
	BaseModel
	WorkspaceID     uuid.UUID `json:"workspace_id"`
	WorkspaceRoleID uuid.UUID `json:"workspace_role_id"`
}

// WorkspaceServiceAccountAccess is a representation of a workspace service account access.
type WorkspaceServiceAccountAccess struct {
	WorkspaceAccessBaseModel
	ActorID uuid.UUID `json:"actor_id"`
	BotID   uuid.UUID `json:"bot_id"`

	BotName string `json:"bot_name"`
}

// WorkspaceUserAccess is a representation of a workspace user access.
type WorkspaceUserAccess struct {
	WorkspaceAccessBaseModel
	ActorID uuid.UUID `json:"actor_id"`
	UserID  uuid.UUID `json:"user_id"`

	Email     string `json:"email"`
	Handle    string `json:"handle"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// WorkspaceAccessUpsert defines the payload
// when upserting a workspace access request.
type WorkspaceAccessUpsert struct {
	WorkspaceRoleID uuid.UUID  `json:"workspace_role_id"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	BotID           *uuid.UUID `json:"bot_id,omitempty"`
}
