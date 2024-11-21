package api

import (
	"context"
	"encoding/json"
	"fmt"

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
type AutomationUpsert struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Enabled          bool     `json:"enabled"`
	Trigger          Trigger  `json:"trigger"`
	Actions          []Action `json:"actions"`
	ActionsOnTrigger []Action `json:"actions_on_trigger"`
	ActionsOnResolve []Action `json:"actions_on_resolve"`
}

type Trigger struct {
	Type string `json:"type"`

	// For EventTrigger
	Match        *ResourceSpecification `json:"match,omitempty"`
	MatchRelated *ResourceSpecification `json:"match_related,omitempty"`
	Posture      *string                `json:"posture,omitempty"`
	After        []string               `json:"after,omitempty"`
	Expect       []string               `json:"expect,omitempty"`
	ForEach      []string               `json:"for_each,omitempty"`
	Threshold    *int                   `json:"threshold,omitempty"`
	Within       *string                `json:"within,omitempty"` // Duration string
	// For MetricTrigger
	Metric *MetricTriggerQuery `json:"metric,omitempty"`
	// For CompoundTrigger
	Triggers []Trigger   `json:"triggers,omitempty"`
	Require  interface{} `json:"require,omitempty"` // int or string ("any"/"all")
}

type MetricTriggerQuery struct {
	Name      string  `json:"name"`
	Threshold float64 `json:"threshold"`
	Operator  string  `json:"operator"` // "<", "<=", ">", ">="
	Range     int     `json:"range"`
	FiringFor int     `json:"firing_for"`
}

// ResourceSpecification is a composite type that returns a map
// where the keys are strings and the values
// can be either (1) a string or (2) a list of strings.
//
// ex:
//
//	{
//	  "resource_type": "aws_s3_bucket",
//	  "tags": ["tag1", "tag2"]
//	}
//
// This is used for the `match` and `match_related` fields.
type ResourceSpecification map[string]StringOrSlice

type StringOrSlice struct {
	String     string
	StringList []string
	IsList     bool
}

// For marshalling a ResourceSpecification to JSON.
func (s StringOrSlice) MarshalJSON() ([]byte, error) {
	var val interface{}
	val = s.StringList
	if !s.IsList {
		val = s.String
	}

	bytes, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("marshal string or slice: %w", err)
	}

	return bytes, nil
}

// For unmarshalling a ResourceSpecification from JSON.
func (s *StringOrSlice) UnmarshalJSON(data []byte) error {
	// Try as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.String = str
		s.IsList = false

		return nil
	}

	// Try as string slice
	var strList []string
	if err := json.Unmarshal(data, &strList); err == nil {
		s.StringList = strList
		s.IsList = true

		return nil
	}

	return fmt.Errorf("ResourceSpecification must be string or string array")
}

type Action struct {
	// On all actions
	Type string `json:"type"`

	// On WorkPoolAction, WorkQueueAction, DeploymentAction, and AutomationAction
	Source *string `json:"source,omitempty"`

	// DeploymentAction fields
	DeploymentID *uuid.UUID `json:"deployment_id,omitempty"`

	// WorkPoolAction fields
	WorkPoolID *uuid.UUID `json:"work_pool_id,omitempty"`

	// WorkQueueAction fields
	WorkQueueID *uuid.UUID `json:"work_queue_id,omitempty"`

	// AutomationAction fields
	AutomationID *uuid.UUID `json:"automation_id,omitempty"`

	// RunDeployment fields
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	JobVariables map[string]interface{} `json:"job_variables,omitempty"`

	// ChangeFlowRunState fields
	Name    *string `json:"name,omitempty"`
	State   *string `json:"state,omitempty"`
	Message *string `json:"message,omitempty"`

	// Webhook fields
	BlockDocumentID *uuid.UUID `json:"block_document_id,omitempty"`
	Payload         *string    `json:"payload,omitempty"`

	// Notification fields
	Subject *string `json:"subject,omitempty"`
	Body    *string `json:"body,omitempty"`
}
