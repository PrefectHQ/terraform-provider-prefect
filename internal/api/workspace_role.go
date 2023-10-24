package api

import (
	"context"

	"github.com/google/uuid"
)

type WorkspaceRolesClient interface {
	Create(ctx context.Context, data WorkspaceRoleUpsert) (*WorkspaceRole, error)
	Update(ctx context.Context, id uuid.UUID, data WorkspaceRoleUpsert) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, roleNames []string) ([]*WorkspaceRole, error)
	Get(ctx context.Context, id uuid.UUID) (*WorkspaceRole, error)
}

// WorkspaceRole is a representation of a workspace role.
type WorkspaceRole struct {
	BaseModel
	Name            string     `json:"name"`
	Description     *string    `json:"description"`
	Permissions     []string   `json:"permissions"`
	Scopes          []string   `json:"scopes"`
	AccountID       *uuid.UUID `json:"account_id"` // this is null for the default roles
	InheritedRoleID *uuid.UUID `json:"inherited_role_id"`
}

// WorkspaceRoleUpsert defines the request payload
// when creating or updating a workspace role.
type WorkspaceRoleUpsert struct {
	Name            string     `json:"name"`
	Description     *string    `json:"description"`
	Scopes          []string   `json:"scopes"`
	InheritedRoleID *uuid.UUID `json:"inherited_role_id"`
}

// WorkspaceRoleFilter defines the search filter payload
// when searching for workspace roles by name.
// example request payload:
// {"workspace_roles": {"name": {"any_": ["test"]}}}.
type WorkspaceRoleFilter struct {
	WorkspaceRoles struct {
		Name struct {
			Any []string `json:"any_"`
		} `json:"name"`
	} `json:"workspace_roles"`
}
