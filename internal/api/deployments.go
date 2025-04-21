package api

import (
	"context"

	"github.com/google/uuid"
)

// DeploymentsClient is a client for working with deployemts.
type DeploymentsClient interface {
	Create(ctx context.Context, data DeploymentCreate) (*Deployment, error)
	Get(ctx context.Context, deploymentID uuid.UUID) (*Deployment, error)
	GetByName(ctx context.Context, flowName, deploymentName string) (*Deployment, error)
	Update(ctx context.Context, deploymentID uuid.UUID, data DeploymentUpdate) error
	Delete(ctx context.Context, deploymentID uuid.UUID) error
}

// Deployment is a representation of a deployment.
type Deployment struct {
	BaseModel
	AccountID   uuid.UUID `json:"account_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`

	ConcurrencyLimit       *int64                         `json:"concurrency_limit"`
	ConcurrencyOptions     *ConcurrencyOptions            `json:"concurrency_options"`
	Description            string                         `json:"description"`
	EnforceParameterSchema bool                           `json:"enforce_parameter_schema"`
	Entrypoint             string                         `json:"entrypoint"`
	FlowID                 uuid.UUID                      `json:"flow_id"`
	GlobalConcurrencyLimit *CurrentGlobalConcurrencyLimit `json:"global_concurrency_limit"`
	JobVariables           map[string]interface{}         `json:"job_variables"`
	Name                   string                         `json:"name"`
	ParameterOpenAPISchema map[string]interface{}         `json:"parameter_openapi_schema"`
	Parameters             map[string]interface{}         `json:"parameters"`
	Path                   string                         `json:"path"`
	Paused                 bool                           `json:"paused"`
	PullSteps              []PullStep                     `json:"pull_steps"`
	StorageDocumentID      uuid.UUID                      `json:"storage_document_id"`
	Tags                   []string                       `json:"tags"`
	Version                string                         `json:"version"`
	WorkPoolName           string                         `json:"work_pool_name"`
	WorkQueueName          string                         `json:"work_queue_name"`
}

// DeploymentCreate is a subset of Deployment used when creating deployments.
type DeploymentCreate struct {
	ConcurrencyLimit       *int64                 `json:"concurrency_limit,omitempty"`
	ConcurrencyOptions     *ConcurrencyOptions    `json:"concurrency_options,omitempty"`
	Description            string                 `json:"description,omitempty"`
	EnforceParameterSchema bool                   `json:"enforce_parameter_schema,omitempty"`
	Entrypoint             string                 `json:"entrypoint,omitempty"`
	FlowID                 uuid.UUID              `json:"flow_id"` // required
	JobVariables           map[string]interface{} `json:"job_variables,omitempty"`
	Name                   string                 `json:"name"` // required
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

// ConcurrencyOptions is a representation of the deployment concurrency options.
type ConcurrencyOptions struct {
	CollisionStrategy string `json:"collision_strategy"`
}

// CurrentGlobalConcurrencyLimit is a representation of the deployment global concurrency limit.
type CurrentGlobalConcurrencyLimit struct {
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

// PullStepCommon is a representation of the common fields for certain pull steps.
type PullStepCommon struct {
	// Credentials is the credentials to use for the pull step.
	// Used on all PullStep types.
	Credentials *string `json:"credentials,omitempty"`

	// Requires is a list of Python package dependencies.
	Requires *string `json:"requires,omitempty"`
}

// PullStepGitClone is a representation of a pull step that clones a git repository.
type PullStepGitClone struct {
	PullStepCommon

	// The URL of the repository to clone.
	Repository *string `json:"repository,omitempty"`

	// The branch to clone. If not provided, the default branch is used.
	Branch *string `json:"branch,omitempty"`

	// Access token for the repository.
	AccessToken *string `json:"access_token,omitempty"`

	// IncludeSubmodules determines whether to include submodules when cloning the repository.
	IncludeSubmodules *bool `json:"include_submodules,omitempty"`
}

// PullStepSetWorkingDirectory is a representation of a pull step that sets the working directory.
type PullStepSetWorkingDirectory struct {
	// The directory to set as the working directory.
	Directory *string `json:"directory,omitempty"`
}

// PullStepPullFrom is a representation of a pull step that pulls from a remote storage bucket.
type PullStepPullFrom struct {
	PullStepCommon

	// The name of the bucket where files are stored.
	Bucket *string `json:"bucket,omitempty"`

	// The folder in the bucket where files are stored.
	Folder *string `json:"folder,omitempty"`
}

// PullStep contains instructions for preparing your flows for a deployment run.
type PullStep struct {
	PullStepGitClone                 *PullStepGitClone            `json:"prefect.deployments.steps.git_clone,omitempty"`
	PullStepSetWorkingDirectory      *PullStepSetWorkingDirectory `json:"prefect.deployments.steps.set_working_directory,omitempty"`
	PullStepPullFromAzureBlobStorage *PullStepPullFrom            `json:"prefect_azure.deployments.steps.pull_from_azure_blob_storage,omitempty"`
	PullStepPullFromGCS              *PullStepPullFrom            `json:"prefect_gcp.deployments.steps.pull_from_gcs,omitempty"`
	PullStepPullFromS3               *PullStepPullFrom            `json:"prefect_aws.deployments.steps.pull_from_s3,omitempty"`
}
