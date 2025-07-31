package api

import (
	"context"
)

// AccountsClient is a client for working with accounts.
type AccountsClient interface {
	Get(ctx context.Context) (*Account, error)
	GetDomains(ctx context.Context) ([]*AccountDomain, error)
	Update(ctx context.Context, data AccountUpdate) error
	UpdateSettings(ctx context.Context, data AccountSettingsUpdate) error
	UpdateDomains(ctx context.Context, data AccountDomainsUpdate) error
	Delete(ctx context.Context) error
}

// AccountSettings is a representation of an account's settings.
type AccountSettings struct {
	AllowPublicWorkspaces bool `json:"allow_public_workspaces"`
	AILogSummaries        bool `json:"ai_log_summaries"`
	ManagedExecution      bool `json:"managed_execution"`
}

// Account is a representation of an account.
type Account struct {
	BaseModel
	AccountUpdate

	Settings AccountSettings `json:"settings"`
	Domains  []string        `json:"domain_names"`

	// Read-only fields
	ImageLocation         *string  `json:"image_location"`
	SSOState              string   `json:"sso_state"`
	Features              []string `json:"features"`
	PlanType              string   `json:"plan_type"`
	RunRetentionDays      int64    `json:"run_retention_days"`
	AuditLogRetentionDays int64    `json:"audit_log_retention_days"`
	AutomationsLimit      int64    `json:"automations_limit"`
	BillingEmail          *string  `json:"billing_email"`
}

// AccountUpdate is the data sent when updating an account.
type AccountUpdate struct {
	Name                  string  `json:"name"`
	Handle                string  `json:"handle"`
	Location              *string `json:"location"`
	Link                  *string `json:"link"`
	AuthExpirationSeconds *int64  `json:"auth_expiration_seconds"`
}

// AccountSettingsUpdate is the data sent when updating an account's settings.
type AccountSettingsUpdate struct {
	AccountSettings `json:"settings"`
}

// AccountDomain is the data retrieved when getting an account's domain names.
type AccountDomain struct {
	Name string `json:"name"`

	// The fields below are present in the response but are currently ignored.
	// id, created, updated, account_id
}

// AccountDomainsUpdate is the data sent when updating an account's domain names.
type AccountDomainsUpdate struct {
	DomainNames []string `json:"domain_names,omitempty"`
}
