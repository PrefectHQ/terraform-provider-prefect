package api

import (
	"context"

	"github.com/google/uuid"
)

// DeploymentsClient is a client for working with deployemts.
type DeploymentsClient interface {
	Create(ctx context.Context, data DeploymentCreate) (*Deployment, error)
	Get(ctx context.Context, deploymentID uuid.UUID) (*Deployment, error)
	List(ctx context.Context, handleNames []string) ([]*Deployment, error)
	Update(ctx context.Context, deploymentID uuid.UUID, data DeploymentUpdate) error
	Delete(ctx context.Context, deploymentID uuid.UUID) error
}

// Deployment is a representation of a deployment.
type Deployment struct {
	BaseModel
	AccountID   uuid.UUID `json:"account_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	// Description            *string   `json:"description"`
	// Handle                 string    `json:"handle"`
	// DefaultWorkspaceRoleID uuid.UUID `json:"default_workspace_role_id"`
	// IsPublic               bool      `json:"is_public"`
}

// DeploymentCreate is a subset of Deployment used when creating deployments.
type DeploymentCreate struct {
	Name string `json:"name"`
	// Description *string `json:"description"`
	// Handle      string  `json:"handle"`
}

// DeploymentUpdate is a subset of Deployment used when updating deployments.
type DeploymentUpdate struct {
	Name *string `json:"name"`
	// Description            *string    `json:"description"`
	// Handle                 *string    `json:"handle"`
	// DefaultWorkspaceRoleID *uuid.UUID `json:"default_workspace_role_id"`
}

// DeploymentFilter defines the search filter payload
// when searching for deployements by name.
// example request payload:
// {"deployments": {"handle": {"any_": ["test"]}}}.
type DeploymentFilter struct {
	Deployments struct {
		Handle struct {
			Any []string `json:"any_"`
		} `json:"handle"`
	} `json:"deployments"`
}
