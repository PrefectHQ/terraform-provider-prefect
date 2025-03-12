package api

import (
	"context"

	"github.com/google/uuid"
)

// SLAsClient is a client for working with SLAs.
type SLAsClient interface {
	ApplyResourceSLAs(ctx context.Context, resourceID uuid.UUID, SLAs []SLAUpsert) (*SLAResponse, error)
}

// SLAResponse is the response from the ApplyResourceSLAs method.
type SLAResponse struct {
	Created []SLA `json:"created,omitempty"`
	Updated []SLA `json:"updated,omitempty"`
	Deleted []SLA `json:"deleted,omitempty"`
}

// SLA is a representation of a service level agreement response.
// NOTE: an SLA response object is a superset of the Automation response object.
type SLA struct {
	Automation
	Severity string `json:"severity"`
	Type     string `json:"type"`
}

// SLAUpsert is the request data needed to create or update an SLA.
type SLAUpsert struct {
	Name          string `json:"name"`
	Severity      string `json:"severity"`
	Enabled       bool   `json:"enabled"`
	OwnerResource string `json:"owner_resource"`

	// For TimeToCompletionSLA
	Duration *float64 `json:"duration,omitempty"`

	// For FrequencySLA
	StaleAfter *float64 `json:"stale_after,omitempty"`

	// For FreshnessSLA
	Within *float64 `json:"within,omitempty"`

	// For LatenessSLA
	After *float64 `json:"after,omitempty"`
}
