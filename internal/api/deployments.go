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
	SetAccess(ctx context.Context, accessControl DeploymentAccessSet) (*DeploymentAccess, error)
	ReadAccess(ctx context.Context, accessControl DeploymentAccessRead) (*DeploymentAccess, error)
}

// Deployment is a representation of a deployment.
type Deployment struct {
	BaseModel
	AccountID              uuid.UUID `json:"account_id"`
	WorkspaceID            uuid.UUID `json:"workspace_id"`
	Name                   string    `json:"name"`
	FlowID                 uuid.UUID `json:"flow_id"`
	IsScheduleActive       bool      `json:"is_schedule_active"`
	Paused                 bool      `json:"paused"`
	EnforceParameterSchema bool      `json:"enforce_parameter_schema"`
	Path                   string    `json:"path"`
	Entrypoint             string    `json:"entrypoint"`
	Tags                   []string  `json:"tags"`

	ManifestPath             string `json:"manifest_path,omitempty"`
	StorageDocumentID        string `json:"storage_document_id,omitempty"`
	InfrastructureDocumentID string `json:"infrastructure_document_id,omitempty"`
	Description              string `json:"description,omitempty"`
	Version                  string `json:"version,omitempty"`

	WorkQueueName string `json:"work_queue_name,omitempty"`
	WorkPoolName  string `json:"work_pool_name,omitempty"`
}

// DeploymentCreate is a subset of Deployment used when creating deployments.
type DeploymentCreate struct {
	Name                   string    `json:"name"`
	FlowID                 uuid.UUID `json:"flow_id"`
	IsScheduleActive       bool      `json:"is_schedule_active"`
	Paused                 bool      `json:"paused"`
	EnforceParameterSchema bool      `json:"enforce_parameter_schema"`
	Path                   string    `json:"path"`
	Entrypoint             string    `json:"entrypoint"`
	Description            string    `json:"description"`
	Tags                   []string  `json:"tags"`
}

// DeploymentUpdate is a subset of Deployment used when updating deployments.
type DeploymentUpdate struct {
	Name *string  `json:"name"`
	Tags []string `json:"tags"`
}

// DeploymentFilter defines the search filter payload
// when searching for deployements by name.
// example request payload:
// {"deployments": {"handle": {"any_": ["test"]}}}.
type DeploymentFilter struct {
}

type DeploymentAccess struct {
	BaseModel
	AccountID     uuid.UUID               `json:"account_id"`
	WorkspaceID   uuid.UUID               `json:"workspace_id"`
	DeploymentID  uuid.UUID               `json:"deployment_id"`
	AccessControl DeploymentAccessControl `json:"access_control"`
}

// DeploymentAccessSet is a subset of DeploymentAccess used when Setting deployment access control.
type DeploymentAccessSet struct {
	DeploymentID  uuid.UUID               `json:"deployment_id"`
	AccessControl DeploymentAccessControl `json:"access_control"`
}

// DeploymentAccessRead is a subset of DeploymentAccess used when Reading deployment access control.
type DeploymentAccessRead struct {
	DeploymentID uuid.UUID `json:"deployment_id"`
}

// DeploymentAccessControl is a defintion of deployment access control.
type DeploymentAccessControl struct {
	ManageActorIDs []string `json:"manage_actor_ids,omitempty"`
	RunActorIDs    []string `json:"run_actor_ids,omitempty"`
	ViewActorIDs   []string `json:"view_actor_ids,omitempty"`
	ManageTeamIDs  []string `json:"manage_team_ids,omitempty"`
	RunTeamIDs     []string `json:"run_team_ids,omitempty"`
	ViewTeamIDs    []string `json:"view_team_ids,omitempty"`
}
