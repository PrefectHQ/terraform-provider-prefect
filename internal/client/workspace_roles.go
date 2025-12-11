package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

// type assertion ensures that this client implements the interface.
var _ = api.WorkspaceRolesClient(&WorkspaceRolesClient{})

type WorkspaceRolesClient struct {
	hc              *http.Client
	apiKey          string
	basicAuthKey    string
	routePrefix     string
	csrfClientToken string
	csrfToken       string
	customHeaders   map[string]string
}

// WorkspaceRoles is a factory that initializes and returns a WorkspaceRolesClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) WorkspaceRoles(accountID uuid.UUID) (api.WorkspaceRolesClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	return &WorkspaceRolesClient{
		hc:              c.hc,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		routePrefix:     getAccountScopedURL(c.endpoint, accountID, "workspace_roles"),
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
	}, nil
}

// Create creates a new workspace role.
func (c *WorkspaceRolesClient) Create(ctx context.Context, data api.WorkspaceRoleUpsert) (*api.WorkspaceRole, error) {
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

	var workspaceRole api.WorkspaceRole
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &workspaceRole); err != nil {
		return nil, fmt.Errorf("failed to create workspace role: %w", err)
	}

	return &workspaceRole, nil
}

// Update modifies an existing workspace role by ID.
func (c *WorkspaceRolesClient) Update(ctx context.Context, workspaceRoleID uuid.UUID, data api.WorkspaceRoleUpsert) error {
	cfg := requestConfig{
		method:          http.MethodPatch,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, workspaceRoleID.String()),
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
		return fmt.Errorf("failed to update workspace role: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Delete removes a workspace role by ID.
func (c *WorkspaceRolesClient) Delete(ctx context.Context, id uuid.UUID) error {
	cfg := requestConfig{
		method:          http.MethodDelete,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, id.String()),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete workspace role: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// List returns a list of workspace roles, based on the provided filter.
func (c *WorkspaceRolesClient) List(ctx context.Context, roleNames []string) ([]*api.WorkspaceRole, error) {
	filterQuery := api.WorkspaceRoleFilter{}
	filterQuery.WorkspaceRoles.Name.Any = roleNames

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

	var workspaceRoles []*api.WorkspaceRole
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &workspaceRoles); err != nil {
		return nil, fmt.Errorf("failed to list workspace roles: %w", err)
	}

	return workspaceRoles, nil
}

// Get returns a workspace role by ID.
func (c *WorkspaceRolesClient) Get(ctx context.Context, id uuid.UUID) (*api.WorkspaceRole, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, id.String()),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusOK,
	}

	var workspaceRole api.WorkspaceRole
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &workspaceRole); err != nil {
		return nil, fmt.Errorf("failed to get workspace role: %w", err)
	}

	return &workspaceRole, nil
}
