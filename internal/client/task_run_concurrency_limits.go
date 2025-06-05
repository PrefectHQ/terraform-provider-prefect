package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.TaskRunConcurrencyLimitsClient(&TaskRunConcurrencyLimitsClient{})

// TaskRunConcurrencyLimitsClient is a client for working with task run concurrency limits.
type TaskRunConcurrencyLimitsClient struct {
	hc              *http.Client
	routePrefix     string
	apiKey          string
	basicAuthKey    string
	csrfClientToken string
	csrfToken       string
}

// TaskRunConcurrencyLimits returns a TaskRunConcurrencyLimitsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) TaskRunConcurrencyLimits(accountID uuid.UUID, workspaceID uuid.UUID) (api.TaskRunConcurrencyLimitsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &TaskRunConcurrencyLimitsClient{
		hc:              c.hc,
		routePrefix:     getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "concurrency_limits"),
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}, nil
}

// Create creates a new task run concurrency limit.
func (c *TaskRunConcurrencyLimitsClient) Create(ctx context.Context, data api.TaskRunConcurrencyLimitCreate) (*api.TaskRunConcurrencyLimit, error) {
	cfg := requestConfig{
		method:          http.MethodPost,
		url:             c.routePrefix + "/",
		body:            &data,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var taskRunConcurrencyLimit api.TaskRunConcurrencyLimit
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &taskRunConcurrencyLimit); err != nil {
		return nil, fmt.Errorf("failed to create task run concurrency limit: %w", err)
	}

	return &taskRunConcurrencyLimit, nil
}

// Read returns a task run concurrency limit.
func (c *TaskRunConcurrencyLimitsClient) Read(ctx context.Context, taskRunConcurrencyLimitID string) (*api.TaskRunConcurrencyLimit, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, taskRunConcurrencyLimitID),
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var taskRunConcurrencyLimit api.TaskRunConcurrencyLimit
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &taskRunConcurrencyLimit); err != nil {
		return nil, fmt.Errorf("failed to get task run concurrency limit: %w", err)
	}

	return &taskRunConcurrencyLimit, nil
}

// Delete deletes a task run concurrency limit.
func (c *TaskRunConcurrencyLimitsClient) Delete(ctx context.Context, taskRunConcurrencyLimitID string) error {
	cfg := requestConfig{
		method:          http.MethodDelete,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, taskRunConcurrencyLimitID),
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete task run concurrency limit: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
