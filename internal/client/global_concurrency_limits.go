package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.GlobalConcurrencyLimitsClient(&GlobalConcurrencyLimitsClient{})

// GlobalConcurrencyLimitsClient is a client for working with global concurrency limits.
type GlobalConcurrencyLimitsClient struct {
	hc              *http.Client
	routePrefix     string
	apiKey          string
	basicAuthKey    string
	csrfClientToken string
	csrfToken       string
	customHeaders   map[string]string
}

// GlobalConcurrencyLimits returns a GlobalConcurrencyLimitsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) GlobalConcurrencyLimits(accountID uuid.UUID, workspaceID uuid.UUID) (api.GlobalConcurrencyLimitsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &GlobalConcurrencyLimitsClient{
		hc:              c.hc,
		routePrefix:     getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "v2/concurrency_limits"),
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
	}, nil
}

// Create creates a new global concurrency limit.
func (c *GlobalConcurrencyLimitsClient) Create(ctx context.Context, data api.GlobalConcurrencyLimitCreate) (*api.GlobalConcurrencyLimit, error) {
	cfg := requestConfig{
		method:          http.MethodPost,
		url:             c.routePrefix + "/",
		body:            &data,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusCreated,
	}

	var globalConcurrencyLimit api.GlobalConcurrencyLimit
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &globalConcurrencyLimit); err != nil {
		return nil, fmt.Errorf("failed to create global concurrency limit: %w", err)
	}

	return &globalConcurrencyLimit, nil
}

// Read returns a global concurrency limit.
func (c *GlobalConcurrencyLimitsClient) Read(ctx context.Context, globalConcurrencyLimitID string) (*api.GlobalConcurrencyLimit, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, globalConcurrencyLimitID),
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusOK,
	}

	var globalConcurrencyLimit api.GlobalConcurrencyLimit
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &globalConcurrencyLimit); err != nil {
		return nil, fmt.Errorf("failed to get global concurrency limit: %w", err)
	}

	return &globalConcurrencyLimit, nil
}

// Update updates a global concurrency limit.
func (c *GlobalConcurrencyLimitsClient) Update(ctx context.Context, globalConcurrencyLimitID string, data api.GlobalConcurrencyLimitUpdate) error {
	cfg := requestConfig{
		method:          http.MethodPatch,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, globalConcurrencyLimitID),
		body:            &data,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update global concurrency limit: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Delete deletes a global concurrency limit.
func (c *GlobalConcurrencyLimitsClient) Delete(ctx context.Context, globalConcurrencyLimitID string) error {
	cfg := requestConfig{
		method:          http.MethodDelete,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, globalConcurrencyLimitID),
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete global concurrency limit: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
