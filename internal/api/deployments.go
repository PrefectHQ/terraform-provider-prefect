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

	ConcurrencyOptions     ConcurrencyOptions     `json:"concurrency_options,omitempty"`
	ConcurrencyLimit       int64                  `json:"concurrency_limit"`
	Description            string                 `json:"description,omitempty"`
	EnforceParameterSchema bool                   `json:"enforce_parameter_schema"`
	Entrypoint             string                 `json:"entrypoint"`
	FlowID                 uuid.UUID              `json:"flow_id"`
	GlobalConcurrencyLimit GlobalConcurrencyLimit `json:"global_concurrency_limit"`
	JobVariables           map[string]interface{} `json:"job_variables,omitempty"`
	ManifestPath           string                 `json:"manifest_path,omitempty"`
	Name                   string                 `json:"name"`
	ParameterOpenAPISchema map[string]interface{} `json:"parameter_openapi_schema,omitempty"`
	Parameters             map[string]interface{} `json:"parameters,omitempty"`
	Path                   string                 `json:"path"`
	Paused                 bool                   `json:"paused"`
	StorageDocumentID      uuid.UUID              `json:"storage_document_id,omitempty"`
	Tags                   []string               `json:"tags"`
	Version                string                 `json:"version,omitempty"`
	WorkPoolName           string                 `json:"work_pool_name,omitempty"`
	WorkQueueName          string                 `json:"work_queue_name,omitempty"`
}

// DeploymentCreate is a subset of Deployment used when creating deployments.
type DeploymentCreate struct {
	ConcurrencyOptions     ConcurrencyOptions     `json:"concurrency_options,omitempty"`
	ConcurrencyLimit       *int64                 `json:"concurrency_limit,omitempty"`
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
	StorageDocumentID      *uuid.UUID             `json:"storage_document_id,omitempty"`
	Tags                   []string               `json:"tags,omitempty"`
	Version                string                 `json:"version,omitempty"`
	WorkPoolName           string                 `json:"work_pool_name,omitempty"`
	WorkQueueName          string                 `json:"work_queue_name,omitempty"`
}

// DeploymentUpdate is a subset of Deployment used when updating deployments.
type DeploymentUpdate struct {
	ConcurrencyOptions     ConcurrencyOptions     `json:"concurrency_options,omitempty"`
	ConcurrencyLimit       *int64                 `json:"concurrency_limit,omitempty"`
	Description            string                 `json:"description,omitempty"`
	EnforceParameterSchema bool                   `json:"enforce_parameter_schema,omitempty"`
	Entrypoint             string                 `json:"entrypoint,omitempty"`
	JobVariables           map[string]interface{} `json:"job_variables,omitempty"`
	ManifestPath           string                 `json:"manifest_path"`
	Parameters             map[string]interface{} `json:"parameters,omitempty"`
	Path                   string                 `json:"path,omitempty"`
	Paused                 bool                   `json:"paused,omitempty"`
	StorageDocumentID      *uuid.UUID             `json:"storage_document_id,omitempty"`
	Tags                   []string               `json:"tags,omitempty"`
	Version                string                 `json:"version,omitempty"`
	WorkPoolName           string                 `json:"work_pool_name,omitempty"`
	WorkQueueName          string                 `json:"work_queue_name,omitempty"`
}

// ConcurrencyOptions is a representation of the deployment concurrency options.
type ConcurrencyOptions struct {
	CollisionStrategy string `json:"collision_strategy,omitempty"`
}

// GlobalConcurrencyLimit is a representation of the deployment global concurrency limit.
type GlobalConcurrencyLimit struct {
	Limit int `json:"limit"`

	// BaseModel
	// Active             bool   `json:"active"`
	// Name               string `json:"name"`
	// ActiveSlots        int    `json:"active_slots"`
	// SlotDecayPerSecond int    `json:"slot_decay_per_second"`
}
