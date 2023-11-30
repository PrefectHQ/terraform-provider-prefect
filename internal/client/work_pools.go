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

var _ = api.WorkPoolsClient(&WorkPoolsClient{})

// WorkPoolsClient is a client for working with work pools.
type WorkPoolsClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}

// WorkPools returns a WorkPoolsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) WorkPools(accountID uuid.UUID, workspaceID uuid.UUID) (api.WorkPoolsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}
	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	return &WorkPoolsClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "work_pools"),
	}, nil
}

// Create returns details for a new work pool.
func (c *WorkPoolsClient) Create(ctx context.Context, data api.WorkPoolCreate) (*api.WorkPool, error) {
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

	var pool api.WorkPool
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pool, nil
}

// List returns a list of work pools matching filter criteria.
func (c *WorkPoolsClient) List(ctx context.Context, filter api.WorkPoolFilter) ([]*api.WorkPool, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&filter); err != nil {
		return nil, fmt.Errorf("failed to encode filter: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.routePrefix+"/filter", &buf)
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

	var pools []*api.WorkPool
	if err := json.NewDecoder(resp.Body).Decode(&pools); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return pools, nil
}

// Get returns details for a work pool by name.
func (c *WorkPoolsClient) Get(ctx context.Context, name string) (*api.WorkPool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.routePrefix+"/"+name, http.NoBody)
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

	var pool api.WorkPool
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pool, nil
}

// Update modifies an existing work pool by name.
func (c *WorkPoolsClient) Update(ctx context.Context, name string, data api.WorkPoolUpdate) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.routePrefix+"/"+name, &buf)
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

// Delete removes a work pool by name.
func (c *WorkPoolsClient) Delete(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.routePrefix+"/"+name, http.NoBody)
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
