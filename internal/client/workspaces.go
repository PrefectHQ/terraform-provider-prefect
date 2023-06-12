package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.WorkspacesClient(&WorkspacesClient{})

// WorkspacesClient is a client for working with Workspaces.
type WorkspacesClient struct {
	hc          *http.Client
	routePrefix string
	apiKey      string
}

// Workspaces returns a WorkspacesClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Workspaces(accountID uuid.UUID) (api.WorkspacesClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	return &WorkspacesClient{
		hc:          c.hc,
		routePrefix: getAccountScopedURL(c.endpoint, accountID, "workspaces"),
		apiKey:      c.apiKey,
	}, nil
}

// Create returns details for a new Workspace.
func (c *WorkspacesClient) Create(ctx context.Context, data api.WorkspaceCreate) (*api.Workspace, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.routePrefix+"/", &buf)
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
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var workspace api.Workspace
	if err := json.NewDecoder(resp.Body).Decode(&workspace); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspace, nil
}

// Get returns details for a Workspace by ID.
func (c *WorkspacesClient) Get(ctx context.Context, workspaceID uuid.UUID) (*api.Workspace, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.routePrefix+"/"+workspaceID.String(), http.NoBody)
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
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var workspace api.Workspace
	if err := json.NewDecoder(resp.Body).Decode(&workspace); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspace, nil
}

// Update modifies an existing Workspace by ID.
func (c *WorkspacesClient) Update(ctx context.Context, workspaceID uuid.UUID, data api.WorkspaceUpdate) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.routePrefix+"/"+workspaceID.String(), &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, c.apiKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code %s", resp.Status)
	}

	return nil
}

// Delete removes a Workspace by ID.
func (c *WorkspacesClient) Delete(ctx context.Context, workspaceID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.routePrefix+"/"+workspaceID.String(), http.NoBody)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, c.apiKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code %s", resp.Status)
	}

	return nil
}
