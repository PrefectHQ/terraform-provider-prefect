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

	// For EventTrigger
	Match        map[string]interface{} `json:"match,omitempty"`
	MatchRelated map[string]interface{} `json:"match_related,omitempty"`
	Posture      *string                `json:"posture,omitempty"`
	After        []string               `json:"after,omitempty"`
	Expect       []string               `json:"expect,omitempty"`
	ForEach      []string               `json:"for_each,omitempty"`
	Threshold    *int64                 `json:"threshold,omitempty"`
	Within       *float64               `json:"within,omitempty"` // Duration string
	// For MetricTrigger
	Metric *MetricTriggerQuery `json:"metric,omitempty"`
	// For CompoundTrigger
	Triggers []Trigger    `json:"triggers,omitempty"`
	Require  *interface{} `json:"require,omitempty"` // int or string ("any"/"all")
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
