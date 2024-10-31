package api

import (
	"context"

	"github.com/google/uuid"
)

// AutomationsClient is a client for working with automations.
type AutomationsClient interface {
	Get(ctx context.Context, id uuid.UUID) (*Automation, error)
	Create(ctx context.Context, data AutomationCreate) (*Automation, error)
	Update(ctx context.Context, id uuid.UUID, data AutomationUpdate) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type TriggerTypes interface {
	// No methods needed - this is just for type union
}

type TriggerBase struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type EventTrigger struct {
	TriggerBase
	After     []string `json:"after"`
	Expect    []string `json:"expect"`
	ForEach   []string `json:"for_each"`
	Posture   string   `json:"posture"`
	Threshold int      `json:"threshold"`
	Within    int      `json:"within"`
}

// Ensure EventTrigger implements TriggerTypes.
var _ TriggerTypes = (*EventTrigger)(nil)

type MetricTriggerOperator string

const (
	LT  MetricTriggerOperator = "<"
	LTE MetricTriggerOperator = "<="
	GT  MetricTriggerOperator = ">"
	GTE MetricTriggerOperator = ">="
)

type PrefectMetric string

type MetricTriggerQuery struct {
	Name      PrefectMetric         `json:"name"`
	Threshold float64               `json:"threshold"`
	Operator  MetricTriggerOperator `json:"operator"`
	Range     int                   `json:"range"`      // duration in seconds, min 300
	FiringFor int                   `json:"firing_for"` // duration in seconds, min 300
}

type MetricTrigger struct {
	TriggerBase
	Posture string             `json:"posture"`
	Metric  MetricTriggerQuery `json:"metric"`
}

// Ensure MetricTrigger implements TriggerTypes.
var _ TriggerTypes = (*MetricTrigger)(nil)

type CompoundTrigger struct {
	TriggerBase
	Triggers []TriggerTypes `json:"triggers"`
	Require  interface{}    `json:"require"` // int or "any"/"all"
	Within   *int           `json:"within,omitempty"`
}

var _ TriggerTypes = (*CompoundTrigger)(nil)

type SequenceTrigger struct {
	TriggerBase
	Triggers []TriggerTypes `json:"triggers"`
	Within   *int           `json:"within,omitempty"`
}

// Ensure SequenceTrigger implements TriggerTypes.
var _ TriggerTypes = (*SequenceTrigger)(nil)

type Action struct{}

type AutomationCore struct {
	Name             string       `json:"name"`
	Description      string       `json:"description"`
	Enabled          bool         `json:"enabled"`
	Trigger          TriggerTypes `json:"trigger"`
	Actions          []Action     `json:"actions"`
	ActionsOnTrigger []Action     `json:"actions_on_trigger"`
	ActionsOnResolve []Action     `json:"actions_on_resolve"`
}

// Automation represents an automation response.
type Automation struct {
	BaseModel
	AutomationCore
	AccountID   uuid.UUID `json:"account"`
	WorkspaceID uuid.UUID `json:"workspace"`
}

// AutomationCreate is the payload for creating automations.
type AutomationCreate struct {
	AutomationCore
	OwnerResource *string `json:"owner_resource"`
}

// AutomationUpdate is the payload for updating automations.
type AutomationUpdate struct {
	AutomationCore
}
