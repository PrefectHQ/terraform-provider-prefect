package api

import (
	"context"
)

// SLAsClient is a client for working with SLAs.
type SLAsClient interface {
	ApplyResourceSLAs(ctx context.Context, resourceID string, slas []SLAUpsert) (*SLAResponse, error)
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
	Name          string  `json:"name"`
	Severity      *string `json:"severity,omitempty"`
	Enabled       *bool   `json:"enabled,omitempty"`
	OwnerResource *string `json:"owner_resource,omitempty"`

	// For TimeToCompletionSLA
	Duration *int64 `json:"duration,omitempty"`

	// For FrequencySLA
	StaleAfter *float64 `json:"stale_after,omitempty"`

	// For FreshnessSLA or LatenessSLA
	Within *float64 `json:"within,omitempty"`

	// For FreshnessSLA
	ExpectedEvent *string                 `json:"expected_event,omitempty"`
	ResourceMatch *map[string]interface{} `json:"resource_match,omitempty"`
}
