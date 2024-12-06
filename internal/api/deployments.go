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

	ConcurrencyLimit       *int64                  `json:"concurrency_limit"`
	ConcurrencyOptions     *ConcurrencyOptions     `json:"concurrency_options,omitempty"`
	Description            string                  `json:"description,omitempty"`
	EnforceParameterSchema bool                    `json:"enforce_parameter_schema"`
	Entrypoint             string                  `json:"entrypoint"`
	FlowID                 uuid.UUID               `json:"flow_id"`
	GlobalConcurrencyLimit *GlobalConcurrencyLimit `json:"global_concurrency_limit"`
	JobVariables           map[string]interface{}  `json:"job_variables,omitempty"`
	ManifestPath           string                  `json:"manifest_path,omitempty"`
	Name                   string                  `json:"name"`
	ParameterOpenAPISchema map[string]interface{}  `json:"parameter_openapi_schema,omitempty"`
	Parameters             map[string]interface{}  `json:"parameters,omitempty"`
	Path                   string                  `json:"path"`
	Paused                 bool                    `json:"paused"`
	PullSteps              []PullStep              `json:"pull_steps,omitempty"`
	StorageDocumentID      uuid.UUID               `json:"storage_document_id,omitempty"`
	Tags                   []string                `json:"tags"`
	Version                string                  `json:"version,omitempty"`
	WorkPoolName           string                  `json:"work_pool_name,omitempty"`
	WorkQueueName          string                  `json:"work_queue_name,omitempty"`
}

// DeploymentCreate is a subset of Deployment used when creating deployments.
type DeploymentCreate struct {
	ConcurrencyLimit       *int64                 `json:"concurrency_limit,omitempty"`
	ConcurrencyOptions     *ConcurrencyOptions    `json:"concurrency_options,omitempty"`
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
	PullSteps              []PullStep             `json:"pull_steps,omitempty"`
	StorageDocumentID      *uuid.UUID             `json:"storage_document_id,omitempty"`
	Tags                   []string               `json:"tags,omitempty"`
	Version                string                 `json:"version,omitempty"`
	WorkPoolName           string                 `json:"work_pool_name,omitempty"`
	WorkQueueName          string                 `json:"work_queue_name,omitempty"`
}

// DeploymentUpdate is a subset of Deployment used when updating deployments.
type DeploymentUpdate struct {
	ConcurrencyLimit       *int64                 `json:"concurrency_limit,omitempty"`
	ConcurrencyOptions     *ConcurrencyOptions    `json:"concurrency_options"`
	Description            string                 `json:"description,omitempty"`
	EnforceParameterSchema bool                   `json:"enforce_parameter_schema,omitempty"`
	Entrypoint             string                 `json:"entrypoint,omitempty"`
	JobVariables           map[string]interface{} `json:"job_variables,omitempty"`
	ManifestPath           string                 `json:"manifest_path"`
	Parameters             map[string]interface{} `json:"parameters,omitempty"`
	Path                   string                 `json:"path,omitempty"`
	Paused                 bool                   `json:"paused,omitempty"`
	PullSteps              []PullStep             `json:"pull_steps,omitempty"`
	StorageDocumentID      *uuid.UUID             `json:"storage_document_id,omitempty"`
	Tags                   []string               `json:"tags,omitempty"`
	Version                string                 `json:"version,omitempty"`
	WorkPoolName           string                 `json:"work_pool_name,omitempty"`
	WorkQueueName          string                 `json:"work_queue_name,omitempty"`
}

// ConcurrencyOptions is a representation of the deployment concurrency options.
type ConcurrencyOptions struct {
	CollisionStrategy string `json:"collision_strategy"`
}

// GlobalConcurrencyLimit is a representation of the deployment global concurrency limit.
type GlobalConcurrencyLimit struct {
	Limit int64 `json:"limit"`

	// These other fields exist in the response payload, but we don't make use of them at the
	// moment, so we'll leave them disabled for now.
	//
	// BaseModel
	// Active             bool   `json:"active"`
	// Name               string `json:"name"`
	// ActiveSlots        int    `json:"active_slots"`
	// SlotDecayPerSecond int    `json:"slot_decay_per_second"`
}

// PullStep contains instructions for preparing your flows for a deployment run.
type PullStep struct {
	// Type is the type of pull step.
	// One of:
	// - set_working_directory
	// - git_clone
	// - pull_from_azure_blob_storage
	// - pull_from_gcs
	// - pull_from_s3
	Type string `json:"type"`

	// Credentials is the credentials to use for the pull step.
	// Used on all PullStep types.
	Credentials *string `json:"credentials,omitempty"`

	// Requires is a list of Python package dependencies.
	Requires *string `json:"requires,omitempty"`

	//
	// Fields for set_working_directory
	//

	// The directory to set as the working directory.
	Directory *string `json:"directory,omitempty"`

	//
	// Fields for git_clone
	//

	// The URL of the repository to clone.
	Repository *string `json:"repository,omitempty"`

	// The branch to clone. If not provided, the default branch is used.
	Branch *string `json:"branch,omitempty"`

	// Access token for the repository.
	AccessToken *string `json:"access_token,omitempty"`

	//
	// Fields for pull_from_{cloud}
	//

	// The name of the bucket where files are stored.
	Bucket *string `json:"bucket,omitempty"`

	// The folder in the bucket where files are stored.
	Folder *string `json:"folder,omitempty"`
}
