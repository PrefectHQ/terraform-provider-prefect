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

var _ = api.VariablesClient(&VariablesClient{})

// VariablesClient is a client for working with variables.
type VariablesClient struct {
	hc          *http.Client
	endpoint    string
	apiKey      string
	accountID   uuid.UUID
	workspaceID uuid.UUID
}

// Variables returns a VariablesClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Variables(accountID uuid.UUID, workspaceID uuid.UUID) (api.VariablesClient, error) {
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

	return &VariablesClient{
		hc:          c.hc,
		endpoint:    c.endpoint,
		apiKey:      c.apiKey,
		accountID:   accountID,
		workspaceID: workspaceID,
	}, nil
}

// Create returns details for a new variable.
func (c *VariablesClient) Create(ctx context.Context, data api.VariableCreate) (*api.Variable, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/variables", &buf)
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

	var variable api.Variable
	if err := json.NewDecoder(resp.Body).Decode(&variable); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &variable, nil
}

// List returns a list of variables matching filter criteria.
func (c *VariablesClient) List(ctx context.Context, filter api.VariableFilter) ([]api.Variable, error) {
	_ = ctx
	_ = filter

	return nil, nil
}

// Get returns details for a variable by ID.
func (c *VariablesClient) Get(ctx context.Context, variableID uuid.UUID) (*api.Variable, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/variables/"+variableID.String(), http.NoBody)
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

	var variable api.Variable
	if err := json.NewDecoder(resp.Body).Decode(&variable); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &variable, nil
}

// GetByName returns details for a variable by name.
func (c *VariablesClient) GetByName(ctx context.Context, name string) (*api.Variable, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/variables/name/"+name, http.NoBody)
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

	var variable api.Variable
	if err := json.NewDecoder(resp.Body).Decode(&variable); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &variable, nil
}

// Update modifies an existing variable by ID.
func (c *VariablesClient) Update(ctx context.Context, variableID uuid.UUID, data api.VariableUpdate) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.endpoint+"/variables/"+variableID.String(), &buf)
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

// Delete removes a variable by ID.
func (c *VariablesClient) Delete(ctx context.Context, variableID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.endpoint+"/variables/"+variableID.String(), http.NoBody)
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
