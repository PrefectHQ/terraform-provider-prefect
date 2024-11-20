package api

import (
	"context"

	"github.com/google/uuid"
)

type DeploymentScheduleClient interface {
	Create(ctx context.Context, deploymentID uuid.UUID, payload []DeploymentSchedulePayload) ([]*DeploymentSchedule, error)
	Read(ctx context.Context, deploymentID uuid.UUID) ([]*DeploymentSchedule, error)
	Update(ctx context.Context, deploymentID uuid.UUID, scheduleID uuid.UUID, payload DeploymentSchedulePayload) error
	Delete(ctx context.Context, deploymentID uuid.UUID, scheduleID uuid.UUID) error
}

type DeploymentSchedule struct {
	BaseModel
	AccountID   uuid.UUID `json:"account_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`

	DeploymentID uuid.UUID `json:"deployment_id"`

	DeploymentSchedulePayload
}

type DeploymentSchedulePayload struct {
	Active           bool    `json:"active,omitempty"`
	MaxScheduledRuns float32 `json:"max_scheduled_runs,omitempty"`

	// Cloud only
	MaxActiveRuns float32 `json:"max_active_runs,omitempty"`
	Catchup       bool    `json:"catchup,omitempty"`

	Schedule Schedule `json:"schedule,omitempty"`
}

type Schedule struct {
	// All schedule kinds specify an interval.
	Timezone string `json:"timezone,omitempty"`

	// Schedule kind: interval
	Interval   float32 `json:"interval,omitempty"`
	AnchorDate string  `json:"anchor_date,omitempty"`

	// Schedule kind: cron
	Cron  string `json:"cron,omitempty"`
	DayOr bool   `json:"day_or,omitempty"`

	// Schedule kind: rrule
	RRule string `json:"rrule,omitempty"`
}
