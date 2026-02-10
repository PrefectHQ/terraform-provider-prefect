package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.WorkspacesClient(&WorkspacesClient{})

// WorkspacesClient is a client for working with Workspaces.
type WorkspacesClient struct {
	hc              *http.Client
	routePrefix     string
	apiKey          string
	basicAuthKey    string
	csrfClientToken string
	csrfToken       string
	customHeaders   map[string]string
}

// Workspaces returns a WorkspacesClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Workspaces(accountID uuid.UUID) (api.WorkspacesClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	return &WorkspacesClient{
		hc:              c.hc,
		routePrefix:     getAccountScopedURL(c.endpoint, accountID, "workspaces"),
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
	}, nil
}

// Create returns details for a new Workspace.
func (c *WorkspacesClient) Create(ctx context.Context, data api.WorkspaceCreate) (*api.Workspace, error) {
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

	var workspace api.Workspace
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &workspace); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	return &workspace, nil
}

// List returns a list of Workspaces, based on the provided list of handle names.
func (c *WorkspacesClient) List(ctx context.Context, handleNames []string) ([]*api.Workspace, error) {
	filterQuery := api.WorkspaceFilter{}

	if len(handleNames) != 0 {
		filterQuery.Workspaces.Handle.Any = handleNames
	}

	cfg := requestConfig{
		method:          http.MethodPost,
		url:             fmt.Sprintf("%s/filter", c.routePrefix),
		body:            &filterQuery,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusOK,
	}

	var workspaces []*api.Workspace
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &workspaces); err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}

	return workspaces, nil
}

// Get returns details for a Workspace by ID.
func (c *WorkspacesClient) Get(ctx context.Context, workspaceID uuid.UUID) (*api.Workspace, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             c.routePrefix + "/" + workspaceID.String(),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusOK,
	}

	var workspace api.Workspace
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &workspace); err != nil {
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	return &workspace, nil
}

// Update modifies an existing Workspace by ID.
func (c *WorkspacesClient) Update(ctx context.Context, workspaceID uuid.UUID, data api.WorkspaceUpdate) error {
	cfg := requestConfig{
		method:          http.MethodPatch,
		url:             c.routePrefix + "/" + workspaceID.String(),
		body:            &data,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Delete removes a Workspace by ID.
func (c *WorkspacesClient) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	cfg := requestConfig{
		method:          http.MethodDelete,
		url:             c.routePrefix + "/" + workspaceID.String(),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	defer resp.Body.Close()

	return nil
}
