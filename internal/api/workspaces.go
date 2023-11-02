package api

import (
	"context"

	"github.com/google/uuid"
)

// WorkspacesClient is a client for working with workspaces.
type WorkspacesClient interface {
	Create(ctx context.Context, data WorkspaceCreate) (*Workspace, error)
	Get(ctx context.Context, workspaceID uuid.UUID) (*Workspace, error)
	List(ctx context.Context, handleNames []string) ([]*Workspace, error)
	Update(ctx context.Context, workspaceID uuid.UUID, data WorkspaceUpdate) error
	Delete(ctx context.Context, workspaceID uuid.UUID) error
}

// Workspace is a representation of a workspace.
type Workspace struct {
	BaseModel
	AccountID              uuid.UUID `json:"account_id"`
	Name                   string    `json:"name"`
	Description            *string   `json:"description"`
	Handle                 string    `json:"handle"`
	DefaultWorkspaceRoleID uuid.UUID `json:"default_workspace_role_id"`
	IsPublic               bool      `json:"is_public"`
}

// WorkspaceCreate is a subset of Workspace used when creating workspaces.
type WorkspaceCreate struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Handle      string  `json:"handle"`
}

// WorkspaceUpdate is a subset of Workspace used when updating workspaces.
type WorkspaceUpdate struct {
	Name                   *string    `json:"name"`
	Description            *string    `json:"description"`
	Handle                 *string    `json:"handle"`
	DefaultWorkspaceRoleID *uuid.UUID `json:"default_workspace_role_id"`
}

// WorkspaceFilter defines the search filter payload
// when searching for workspaces by name.
// example request payload:
// {"workspaces": {"handle": {"any_": ["test"]}}}.
type WorkspaceFilter struct {
	Workspaces struct {
		Handle struct {
			Any []string `json:"any_"`
		} `json:"handle"`
	} `json:"workspaces"`
}
