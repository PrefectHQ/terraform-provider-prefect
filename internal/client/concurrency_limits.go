package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.ConcurrencyLimitsClient(&ConcurrencyLimitsClient{})

// ConcurrencyLimitsClient is a client for working with concurrency limits.
type ConcurrencyLimitsClient struct {
	hc          *http.Client
	routePrefix string
	apiKey      string
}

// ConcurrencyLimits returns a ConcurrencyLimitsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) ConcurrencyLimits(accountID uuid.UUID, workspaceID uuid.UUID) (api.ConcurrencyLimitsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &ConcurrencyLimitsClient{
		hc:          c.hc,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "concurrency_limits"),
		apiKey:      c.apiKey,
	}, nil
}

// Create creates a new concurrency limit.
func (c *ConcurrencyLimitsClient) Create(ctx context.Context, data api.ConcurrencyLimitCreate) (*api.ConcurrencyLimit, error) {
	cfg := requestConfig{
		method:       http.MethodPost,
		url:          c.routePrefix,
		body:         &data,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusCreated,
	}

	var concurrencyLimit api.ConcurrencyLimit
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &concurrencyLimit); err != nil {
		return nil, fmt.Errorf("failed to create concurrency limit: %w", err)
	}

	return &concurrencyLimit, nil
}

// Get returns a concurrency limit.
func (c *ConcurrencyLimitsClient) Get(ctx context.Context, concurrencyLimitID uuid.UUID) (*api.ConcurrencyLimit, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, concurrencyLimitID.String()),
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOK,
	}

	var concurrencyLimit api.ConcurrencyLimit
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &concurrencyLimit); err != nil {
		return nil, fmt.Errorf("failed to get concurrency limit: %w", err)
	}

	return &concurrencyLimit, nil
}

// GetByTag returns a concurrency limit by tag.
func (c *ConcurrencyLimitsClient) GetByTag(ctx context.Context, tag string) (*api.ConcurrencyLimit, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          fmt.Sprintf("%s/tag/%s", c.routePrefix, tag),
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOK,
	}

	var concurrencyLimit api.ConcurrencyLimit
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &concurrencyLimit); err != nil {
		return nil, fmt.Errorf("failed to get concurrency limit by tag: %w", err)
	}

	return &concurrencyLimit, nil
}

// Delete deletes a concurrency limit.
func (c *ConcurrencyLimitsClient) Delete(ctx context.Context, concurrencyLimitID uuid.UUID) error {
	cfg := requestConfig{
		method:       http.MethodDelete,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, concurrencyLimitID.String()),
		apiKey:       c.apiKey,
		successCodes: successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete concurrency limit: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// DeleteByTag deletes a concurrency limit by tag.
func (c *ConcurrencyLimitsClient) DeleteByTag(ctx context.Context, tag string) error {
	cfg := requestConfig{
		method:       http.MethodDelete,
		url:          fmt.Sprintf("%s/tag/%s", c.routePrefix, tag),
		apiKey:       c.apiKey,
		successCodes: successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete concurrency limit by tag: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
