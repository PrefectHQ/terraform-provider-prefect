package api

import (
	"context"

	"github.com/google/uuid"
)

// WorkPoolAccessClient is a client for working with work pool access.
type WorkPoolAccessClient interface {
	Read(ctx context.Context, workPoolName string) (*WorkPoolAccessControl, error)
	Set(ctx context.Context, workPoolName string, accessControl WorkPoolAccessSet) error
}

// WorkPoolAccess is a representation of a work pool access.
type WorkPoolAccess struct {
	BaseModel
	AccountID     uuid.UUID             `json:"account_id"`
	WorkspaceID   uuid.UUID             `json:"workspace_id"`
	WorkPoolName  string                `json:"work_pool_name"`
	AccessControl WorkPoolAccessControl `json:"access_control"`
}

// WorkPoolAccessSet is a subset of WorkPoolAccess used when setting work pool access control.
type WorkPoolAccessSet struct {
	AccessControl WorkPoolAccessControlSet `json:"access_control"`
}

// WorkPoolAccessControlSet is a definition of work pool access control.
type WorkPoolAccessControlSet struct {
	ManageActorIDs []string `json:"manage_actor_ids"`
	RunActorIDs    []string `json:"run_actor_ids"`
	ViewActorIDs   []string `json:"view_actor_ids"`
	ManageTeamIDs  []string `json:"manage_team_ids"`
	RunTeamIDs     []string `json:"run_team_ids"`
	ViewTeamIDs    []string `json:"view_team_ids"`
}

// WorkPoolAccessControl is a representation of a work pool access control.
type WorkPoolAccessControl struct {
	ManageActors []ObjectActorAccess `json:"manage_actors"`
	RunActors    []ObjectActorAccess `json:"run_actors"`
	ViewActors   []ObjectActorAccess `json:"view_actors"`
}
