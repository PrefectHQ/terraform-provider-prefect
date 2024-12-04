package api

import (
	"context"

	"github.com/google/uuid"
)

// AutomationsClient is a client for working with automations.
type AutomationsClient interface {
	Get(ctx context.Context, id uuid.UUID) (*Automation, error)
	Create(ctx context.Context, data AutomationUpsert) (*Automation, error)
	Update(ctx context.Context, id uuid.UUID, data AutomationUpsert) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// Automation represents an automation response.
type Automation struct {
	BaseModel
	AutomationUpsert
	AccountID   uuid.UUID `json:"account_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
}

// AutomationUpsert is the data needed to create or update an automation.
type AutomationUpsert struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Enabled          bool     `json:"enabled"`
	Trigger          Trigger  `json:"trigger"`
	Actions          []Action `json:"actions"`
	ActionsOnTrigger []Action `json:"actions_on_trigger"`
	ActionsOnResolve []Action `json:"actions_on_resolve"`
}

// Trigger defines the triggering conditions on an Automation.
// On the API, a Trigger is a polymorphic type and can be represented
// by several schemas based on the `type` attribute.
// Here, we'll combine all possible attributes for each type
// and make them optional.
type Trigger struct {
	Type string `json:"type"`

	// For EventTrigger and MetricTrigger
	Match        map[string]interface{} `json:"match,omitempty"`
	MatchRelated map[string]interface{} `json:"match_related,omitempty"`
	Posture      *string                `json:"posture,omitempty"`

	// For EventTrigger
	After     []string `json:"after,omitempty"`
	Expect    []string `json:"expect,omitempty"`
	ForEach   []string `json:"for_each,omitempty"`
	Threshold *int64   `json:"threshold,omitempty"`
	Within    *float64 `json:"within,omitempty"`

	// For MetricTrigger
	Metric *MetricTriggerQuery `json:"metric,omitempty"`

	// For CompoundTrigger and SequenceTrigger
	Triggers []Trigger    `json:"triggers,omitempty"`
	Require  *interface{} `json:"require,omitempty"` // int or string ("any"/"all")
}

type MetricTriggerQuery struct {
	Name      string  `json:"name"`
	Threshold float64 `json:"threshold"`
	Operator  string  `json:"operator"` // "<", "<=", ">", ">="
	Range     float64 `json:"range"`
	FiringFor float64 `json:"firing_for"`
}

// Action defines the actions that can be taken on an Automation.
// On the API, an Action is a polymorphic type and can be represented
// by several schemas based on the `type` attribute.
// Here, we'll combine all possible attributes for each type
// and make them optional.
type Action struct {
	// On all actions
	Type string `json:"type"`

	// On Deployment, Work Pool, Work Queue, and Automation actions
	Source *string `json:"source,omitempty"`

	// On Automation actions
	AutomationID *uuid.UUID `json:"automation_id,omitempty"`

	// On Webhook and Notification actions
	BlockDocumentID *uuid.UUID `json:"block_document_id,omitempty"`

	// On Deployment actions
	DeploymentID *uuid.UUID `json:"deployment_id,omitempty"`

	// On Work Pool actions
	WorkPoolID *uuid.UUID `json:"work_pool_id,omitempty"`

	// On Work Queue actions
	WorkQueueID *uuid.UUID `json:"work_queue_id,omitempty"`

	// On Run Deployment action
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	JobVariables map[string]interface{} `json:"job_variables,omitempty"`

	// On Send Notification action
	Subject *string `json:"subject,omitempty"`
	Body    *string `json:"body,omitempty"`

	// On Call Webhook action
	Payload *string `json:"payload,omitempty"`

	// On Change Flow Run State action
	Name    *string `json:"name,omitempty"`
	State   *string `json:"state,omitempty"`
	Message *string `json:"message,omitempty"`
}
