package api

import (
	"context"

	"github.com/google/uuid"
)

// DeploymentAccessClient is a client for working with deployment access.
type DeploymentAccessClient interface {
	Read(ctx context.Context, deploymentID uuid.UUID) (*DeploymentAccessControl, error)
	Set(ctx context.Context, deploymentID uuid.UUID, accessControl DeploymentAccessSet) error
}

// DeploymentAccess is a representation of a deployment access.
type DeploymentAccess struct {
	BaseModel
	AccountID     uuid.UUID               `json:"account_id"`
	WorkspaceID   uuid.UUID               `json:"workspace_id"`
	DeploymentID  uuid.UUID               `json:"deployment_id"`
	AccessControl DeploymentAccessControl `json:"access_control"`
}

// DeploymentAccessSet is a subset of DeploymentAccess used when setting deployment access control.
type DeploymentAccessSet struct {
	AccessControl DeploymentAccessControlSet `json:"access_control"`
}

// DeploymentAccessControlSet is a definition of deployment access control.
type DeploymentAccessControlSet struct {
	ManageActorIDs []string `json:"manage_actor_ids"`
	RunActorIDs    []string `json:"run_actor_ids"`
	ViewActorIDs   []string `json:"view_actor_ids"`
	ManageTeamIDs  []string `json:"manage_team_ids"`
	RunTeamIDs     []string `json:"run_team_ids"`
	ViewTeamIDs    []string `json:"view_team_ids"`
}

// DeploymentAccessControl is a definition of deployment access control.
type DeploymentAccessControl struct {
	ManageActors []Actor `json:"manage_actors"`
	RunActors    []Actor `json:"run_actors"`
	ViewActors   []Actor `json:"view_actors"`
}

// Actor is a definition of an actor.
type Actor struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Type  string    `json:"type"`
}
