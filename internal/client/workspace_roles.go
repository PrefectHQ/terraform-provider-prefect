package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

// type assertion ensures that this client implements the interface.
var _ = api.WorkspaceRolesClient(&WorkspaceRolesClient{})

type WorkspaceRolesClient struct {
	hc          *retryablehttp.Client
	apiKey      string
	routePrefix string
}

// WorkspaceRoles is a factory that initializes and returns a WorkspaceRolesClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) WorkspaceRoles(accountID uuid.UUID) (api.WorkspaceRolesClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	return &WorkspaceRolesClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: getAccountScopedURL(c.endpoint, accountID, "workspace_roles"),
	}, nil
}

// Create creates a new workspace role.
func (c *WorkspaceRolesClient) Create(ctx context.Context, data api.WorkspaceRoleUpsert) (*api.WorkspaceRole, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return nil, fmt.Errorf("failed to encode create payload data: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/", c.routePrefix), &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, c.apiKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errorBody, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	var workspaceRole api.WorkspaceRole
	if err := json.NewDecoder(resp.Body).Decode(&workspaceRole); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspaceRole, nil
}

// Update modifies an existing workspace role by ID.
func (c *WorkspaceRolesClient) Update(ctx context.Context, workspaceRoleID uuid.UUID, data api.WorkspaceRoleUpsert) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return fmt.Errorf("failed to encode update payload data: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", c.routePrefix, workspaceRoleID.String()), &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, c.apiKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		errorBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	return nil
}

// Delete removes a workspace role by ID.
func (c *WorkspaceRolesClient) Delete(ctx context.Context, id uuid.UUID) error {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", c.routePrefix, id.String()), http.NoBody)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, c.apiKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		errorBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	return nil
}

// List returns a list of workspace roles, based on the provided filter.
func (c *WorkspaceRolesClient) List(ctx context.Context, roleNames []string) ([]*api.WorkspaceRole, error) {
	var buf bytes.Buffer
	filterQuery := api.WorkspaceRoleFilter{}
	filterQuery.WorkspaceRoles.Name.Any = roleNames

	if err := json.NewEncoder(&buf).Encode(&filterQuery); err != nil {
		return nil, fmt.Errorf("failed to encode filter payload data: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/filter", c.routePrefix), &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, c.apiKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	var workspaceRoles []*api.WorkspaceRole
	if err := json.NewDecoder(resp.Body).Decode(&workspaceRoles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return workspaceRoles, nil
}

// Get returns a workspace role by ID.
func (c *WorkspaceRolesClient) Get(ctx context.Context, id uuid.UUID) (*api.WorkspaceRole, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", c.routePrefix, id.String()), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, c.apiKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	var workspaceRole api.WorkspaceRole
	if err := json.NewDecoder(resp.Body).Decode(&workspaceRole); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspaceRole, nil
}
