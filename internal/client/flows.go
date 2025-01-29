package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.FlowsClient(&FlowsClient{})

// FlowsClient is a client for working with Flows.
type FlowsClient struct {
	hc          *http.Client
	routePrefix string
	apiKey      string
}

// Flows returns a FlowsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Flows(accountID uuid.UUID, workspaceID uuid.UUID) (api.FlowsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &FlowsClient{
		hc:          c.hc,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "flows"),
		apiKey:      c.apiKey,
	}, nil
}

// Create returns details for a new Flow.
func (c *FlowsClient) Create(ctx context.Context, data api.FlowCreate) (*api.Flow, error) {
	cfg := requestConfig{
		method:       http.MethodPost + "/",
		url:          c.routePrefix,
		body:         &data,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusCreated,
	}

	var flow api.Flow
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &flow); err != nil {
		return nil, fmt.Errorf("failed to create flow: %w", err)
	}

	return &flow, nil
}

// List returns a list of Flows, based on the provided list of handle names.
func (c *FlowsClient) List(ctx context.Context, handleNames []string) ([]*api.Flow, error) {
	filterQuery := api.WorkspaceFilter{}

	if len(handleNames) != 0 {
		filterQuery.Workspaces.Handle.Any = handleNames
	}

	cfg := requestConfig{
		method:       http.MethodPost,
		url:          fmt.Sprintf("%s/filter", c.routePrefix),
		body:         &filterQuery,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOK,
	}

	var flows []*api.Flow
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &flows); err != nil {
		return nil, fmt.Errorf("failed to list flows: %w", err)
	}

	return flows, nil
}

// Get returns details for a Flow by ID.
func (c *FlowsClient) Get(ctx context.Context, flowID uuid.UUID) (*api.Flow, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, flowID.String()),
		body:         http.NoBody,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOK,
	}

	var flow api.Flow
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &flow); err != nil {
		return nil, fmt.Errorf("failed to get flow: %w", err)
	}

	return &flow, nil
}

// Update modifies an existing Flow by ID.
func (c *FlowsClient) Update(ctx context.Context, flowID uuid.UUID, data api.FlowUpdate) error {
	cfg := requestConfig{
		method:       http.MethodPatch,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, flowID.String()),
		body:         &data,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update flow: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Delete removes a Flow by ID.
func (c *FlowsClient) Delete(ctx context.Context, flowID uuid.UUID) error {
	cfg := requestConfig{
		method:       http.MethodDelete,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, flowID.String()),
		body:         http.NoBody,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete flow: %w", err)
	}

	defer resp.Body.Close()

	return nil
}
