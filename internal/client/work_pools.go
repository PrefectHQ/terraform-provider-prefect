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

var _ = api.WorkPoolsClient(&WorkPoolsClient{})

// WorkPoolsClient is a client for working with work pools.
type WorkPoolsClient struct {
	hc          *http.Client
	endpoint    string
	apiKey      string
	accountID   uuid.UUID
	workspaceID uuid.UUID
}

// WorkPools returns a WorkPoolsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) WorkPools(accountID uuid.UUID, workspaceID uuid.UUID) (api.WorkPoolsClient, error) {
	if accountID != uuid.Nil && workspaceID == uuid.Nil {
		return nil, fmt.Errorf("accountID and workspaceID are inconsistent: accountID is %q and workspaceID is nil", accountID)
	}

	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if accountID != uuid.Nil && workspaceID == uuid.Nil {
		if c.defaultWorkspaceID == uuid.Nil {
			return nil, fmt.Errorf("accountID and workspaceID are inconsistent: accountID is %q and supplied/default workspaceID are both nil", accountID)
		}

		workspaceID = c.defaultWorkspaceID
	}

	return &WorkPoolsClient{
		hc:          c.hc,
		endpoint:    c.endpoint,
		apiKey:      c.apiKey,
		accountID:   accountID,
		workspaceID: workspaceID,
	}, nil
}

// Create returns details for a new work pool.
func (c *WorkPoolsClient) Create(ctx context.Context, data api.WorkPoolCreate) (*api.WorkPool, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/work_pools/", &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := doRequest(c.hc, c.apiKey, req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("status code %s", resp.Status)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/work_pools/filter", &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := doRequest(c.hc, c.apiKey, req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var pools []*api.WorkPool
	if err := json.NewDecoder(resp.Body).Decode(&pools); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return pools, nil
}

// Get returns details for a work pool by name.
func (c *WorkPoolsClient) Get(ctx context.Context, name string) (*api.WorkPool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/work_pools/"+name, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := doRequest(c.hc, c.apiKey, req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %s", resp.Status)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.endpoint+"/work_pools/"+name, &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := doRequest(c.hc, c.apiKey, req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code %s", resp.Status)
	}

	return nil
}

// Delete removes a work pool by name.
func (c *WorkPoolsClient) Delete(ctx context.Context, name string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.endpoint+"/work_pools/"+name, http.NoBody)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	resp, err := doRequest(c.hc, c.apiKey, req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code %s", resp.Status)
	}

	return nil
}
