package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.WorkQueuesClient(&WorkQueuesClient{})

// WorkQueuesClient is a client for working with work queues.
type WorkQueuesClient struct {
	hc           *http.Client
	apiKey       string
	basicAuthKey string
	routePrefix  string
}

// WorkQueues returns a WorkQueuesClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) WorkQueues(accountID uuid.UUID, workspaceID uuid.UUID, workPoolName string) (api.WorkQueuesClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	route := fmt.Sprintf("work_pools/%s/queues", workPoolName)

	return &WorkQueuesClient{
		hc:           c.hc,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		routePrefix:  getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, route),
	}, nil
}

// Create returns details for a new work queue.
func (c *WorkQueuesClient) Create(ctx context.Context, data api.WorkQueueCreate) (*api.WorkQueue, error) {
	cfg := requestConfig{
		method:       http.MethodPost,
		url:          c.routePrefix + "/",
		body:         &data,
		successCodes: successCodesStatusCreated,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
	}

	var queue api.WorkQueue
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &queue); err != nil {
		return nil, fmt.Errorf("failed to create work queue: %w", err)
	}

	return &queue, nil
}

// List returns a list of work queues matching filter criteria.
func (c *WorkQueuesClient) List(ctx context.Context, filter api.WorkQueueFilter) ([]*api.WorkQueue, error) {
	cfg := requestConfig{
		method:       http.MethodPost,
		url:          c.routePrefix + "/filter",
		body:         &filter,
		successCodes: successCodesStatusOK,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
	}

	var queues []*api.WorkQueue
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &queues); err != nil {
		return nil, fmt.Errorf("failed to list work queues: %w", err)
	}

	return queues, nil
}

// Get returns details for a work queue by name.
func (c *WorkQueuesClient) Get(ctx context.Context, name string) (*api.WorkQueue, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          c.routePrefix + "/" + name,
		successCodes: successCodesStatusOK,
		body:         http.NoBody,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
	}

	var queue api.WorkQueue
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &queue); err != nil {
		return nil, fmt.Errorf("failed to get work queue: %w", err)
	}

	return &queue, nil
}

// Update modifies an existing work queue by name.
func (c *WorkQueuesClient) Update(ctx context.Context, name string, data api.WorkQueueUpdate) error {
	cfg := requestConfig{
		method:       http.MethodPatch,
		url:          c.routePrefix + "/" + name,
		body:         &data,
		successCodes: successCodesStatusOKOrNoContent,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update work queue: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Delete removes a work queue by name.
func (c *WorkQueuesClient) Delete(ctx context.Context, name string) error {
	cfg := requestConfig{
		method:       http.MethodDelete,
		url:          c.routePrefix + "/" + name,
		successCodes: successCodesStatusOKOrNoContent,
		body:         http.NoBody,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete work queue: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
