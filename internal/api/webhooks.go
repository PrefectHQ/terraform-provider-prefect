package api

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type WebhooksClient interface {
	Create(ctx context.Context, request WebhookCreateRequest) (*Webhook, error)
	Get(ctx context.Context, webhookID string) (*Webhook, error)
	List(ctx context.Context, names []string) ([]*Webhook, error)
	Update(ctx context.Context, webhookID string, request WebhookUpdateRequest) error
	Delete(ctx context.Context, webhookID string) error
}

/*** REQUEST DATA STRUCTS ***/

type WebhookCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
	Template    string `json:"template"`
}

type WebhookUpdateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Enabled     bool   `json:"enabled"`
	Template    string `json:"template"`
}

/*** RESPONSE DATA STRUCTS ***/

type Webhook struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Enabled     bool      `json:"enabled"`
	Template    string    `json:"template"`
	Created     time.Time `json:"created"`
	Updated     time.Time `json:"updated"`
	AccountID   uuid.UUID `json:"account"`
	WorkspaceID uuid.UUID `json:"workspace"`
	Slug        string    `json:"slug"`
}

type ErrorResponse struct {
	Detail []ErrorDetail `json:"detail"`
}

type ErrorDetail struct {
	Loc  []string `json:"loc"`
	Msg  string   `json:"msg"`
	Type string   `json:"type"`
}

// WebhookFilter defines filters when searching for webhooks.
type WebhookFilter struct {
	Webhooks struct {
		Name struct {
			Any []string `json:"any_"`
		} `json:"name,omitempty"`
	} `json:"webhooks"`
}
