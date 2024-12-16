package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.AutomationsClient(&AutomationsClient{})

type AutomationsClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}

// Automations is a factory that initializes and returns a AutomationsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Automations(accountID uuid.UUID, workspaceID uuid.UUID) (api.AutomationsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &AutomationsClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "automations"),
	}, nil
}

func (c *AutomationsClient) Get(ctx context.Context, id uuid.UUID) (*api.Automation, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, id),
		body:         http.NoBody,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOK,
	}

	var automation api.Automation
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &automation); err != nil {
		return nil, fmt.Errorf("failed to get automation: %w", err)
	}

	return &automation, nil
}

func (c *AutomationsClient) Create(ctx context.Context, payload api.AutomationUpsert) (*api.Automation, error) {
	cfg := requestConfig{
		method:       http.MethodPost,
		url:          c.routePrefix + "/",
		body:         payload,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusCreated,
	}

	var automation api.Automation
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &automation); err != nil {
		return nil, fmt.Errorf("failed to create automation: %w", err)
	}

	return &automation, nil
}

func (c *AutomationsClient) Update(ctx context.Context, id uuid.UUID, payload api.AutomationUpsert) error {
	cfg := requestConfig{
		method:       http.MethodPut,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, id),
		body:         payload,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update automation: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (c *AutomationsClient) Delete(ctx context.Context, id uuid.UUID) error {
	cfg := requestConfig{
		method:       http.MethodDelete,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, id),
		body:         http.NoBody,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete automation: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
