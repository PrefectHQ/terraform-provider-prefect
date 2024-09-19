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

	Description            string                 `json:"description,omitempty"`
	EnforceParameterSchema bool                   `json:"enforce_parameter_schema"`
	Entrypoint             string                 `json:"entrypoint"`
	FlowID                 uuid.UUID              `json:"flow_id"`
	JobVariables           map[string]interface{} `json:"job_variables,omitempty"`
	ManifestPath           string                 `json:"manifest_path,omitempty"`
	Name                   string                 `json:"name"`
	ParameterOpenAPISchema map[string]interface{} `json:"parameter_openapi_schema,omitempty"`
	Parameters             map[string]interface{} `json:"parameters,omitempty"`
	Path                   string                 `json:"path"`
	Paused                 bool                   `json:"paused"`
	Tags                   []string               `json:"tags"`
	Version                string                 `json:"version,omitempty"`
	WorkPoolName           string                 `json:"work_pool_name,omitempty"`
	WorkQueueName          string                 `json:"work_queue_name,omitempty"`
}

// DeploymentCreate is a subset of Deployment used when creating deployments.
type DeploymentCreate struct {
	Description            string                 `json:"description,omitempty"`
	EnforceParameterSchema bool                   `json:"enforce_parameter_schema,omitempty"`
	Entrypoint             string                 `json:"entrypoint,omitempty"`
	FlowID                 uuid.UUID              `json:"flow_id"`
	JobVariables           map[string]interface{} `json:"job_variables,omitempty"`
	ManifestPath           string                 `json:"manifest_path,omitempty"`
	Name                   string                 `json:"name"`
	ParameterOpenAPISchema map[string]interface{} `json:"parameter_openapi_schema,omitempty"`
	Parameters             map[string]interface{} `json:"parameters,omitempty"`
	Path                   string                 `json:"path,omitempty"`
	Paused                 bool                   `json:"paused,omitempty"`
	Tags                   []string               `json:"tags,omitempty"`
	Version                string                 `json:"version,omitempty"`
	WorkPoolName           string                 `json:"work_pool_name,omitempty"`
	WorkQueueName          string                 `json:"work_queue_name,omitempty"`
}

// DeploymentUpdate is a subset of Deployment used when updating deployments.
type DeploymentUpdate struct {
	Description            string                 `json:"description,omitempty"`
	EnforceParameterSchema bool                   `json:"enforce_parameter_schema,omitempty"`
	Entrypoint             string                 `json:"entrypoint,omitempty"`
	JobVariables           map[string]interface{} `json:"job_variables,omitempty"`
	ManifestPath           string                 `json:"manifest_path"`
	Parameters             map[string]interface{} `json:"parameters,omitempty"`
	Path                   string                 `json:"path,omitempty"`
	Paused                 bool                   `json:"paused,omitempty"`
	Tags                   []string               `json:"tags,omitempty"`
	Version                string                 `json:"version,omitempty"`
	WorkPoolName           string                 `json:"work_pool_name,omitempty"`
	WorkQueueName          string                 `json:"work_queue_name,omitempty"`
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

type Actor struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Type  string    `json:"type"`
}
