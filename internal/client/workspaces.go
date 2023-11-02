package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		errorBody, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	var workspace api.Workspace
	if err := json.NewDecoder(resp.Body).Decode(&workspace); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspace, nil
}

// List returns a list of Workspaces, based on the provided list of handle names.
func (c *WorkspacesClient) List(ctx context.Context, handleNames []string) ([]*api.Workspace, error) {
	var buf bytes.Buffer
	filterQuery := api.WorkspaceFilter{}
	filterQuery.Workspaces.Handle.Any = handleNames

	if err := json.NewEncoder(&buf).Encode(&filterQuery); err != nil {
		return nil, fmt.Errorf("failed to encode filter payload data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/filter", c.routePrefix), &buf)
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

	var workspaces []*api.Workspace
	if err := json.NewDecoder(resp.Body).Decode(&workspaces); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return workspaces, nil
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
		errorBody, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
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

	endpoint := c.routePrefix + "/" + workspaceID.String()
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, &buf)
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
		errorBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
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
		errorBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	return nil
}
