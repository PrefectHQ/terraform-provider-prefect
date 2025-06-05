package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.WebhooksClient(&WebhooksClient{})

// WebhooksClient is a client for working with webhooks.
type WebhooksClient struct {
	hc              *http.Client
	apiKey          string
	basicAuthKey    string
	routePrefix     string
	csrfClientToken string
	csrfToken       string
}

// Webhooks returns a WebhooksClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Webhooks(accountID, workspaceID uuid.UUID) (api.WebhooksClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &WebhooksClient{
		hc:              c.hc,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		routePrefix:     getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "webhooks"),
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}, nil
}

// Create creates a new webhook.
func (c *WebhooksClient) Create(ctx context.Context, createPayload api.WebhookCreateRequest) (*api.Webhook, error) {
	cfg := requestConfig{
		method:          http.MethodPost,
		url:             c.routePrefix + "/",
		body:            &createPayload,
		successCodes:    successCodesStatusCreated,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}

	var webhook api.Webhook
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &webhook); err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}

	return &webhook, nil
}

// Get returns details for a webhook by ID.
func (c *WebhooksClient) Get(ctx context.Context, webhookID string) (*api.Webhook, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             c.routePrefix + "/" + webhookID,
		successCodes:    successCodesStatusOK,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}

	var webhook api.Webhook
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &webhook); err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}

	return &webhook, nil
}

// Update modifies an existing webhook by ID.
func (c *WebhooksClient) Update(ctx context.Context, webhookID string, updatePayload api.WebhookUpdateRequest) error {
	cfg := requestConfig{
		method:          http.MethodPut,
		url:             c.routePrefix + "/" + webhookID,
		body:            &updatePayload,
		successCodes:    successCodesStatusOKOrNoContent,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Delete removes a webhook by ID.
func (c *WebhooksClient) Delete(ctx context.Context, webhookID string) error {
	cfg := requestConfig{
		method:          http.MethodDelete,
		url:             c.routePrefix + "/" + webhookID,
		successCodes:    successCodesStatusOKOrNoContent,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// List returns a list of webhooks matching filter criteria.
func (c *WebhooksClient) List(ctx context.Context, names []string) ([]*api.Webhook, error) {
	filter := api.WebhookFilter{}
	filter.Webhooks.Name.Any = names

	cfg := requestConfig{
		method:          http.MethodGet,
		url:             c.routePrefix + "/",
		body:            http.NoBody,
		successCodes:    successCodesStatusOK,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}

	var webhooks []*api.Webhook
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &webhooks); err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}

	return webhooks, nil
}
