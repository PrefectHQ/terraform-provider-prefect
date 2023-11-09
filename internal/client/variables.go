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

var _ = api.VariablesClient(&VariablesClient{})

// VariablesClient is a client for working with variables.
type VariablesClient struct {
	hc          *http.Client
	routePrefix string
	apiKey      string
}

// Variables returns a VariablesClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Variables(accountID uuid.UUID, workspaceID uuid.UUID) (api.VariablesClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}
	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if accountID != uuid.Nil || workspaceID != uuid.Nil {
		return nil, fmt.Errorf("both accountID and workspaceID must be set: accountID is %q and workspaceID is %q", accountID, workspaceID)
	}

	return &VariablesClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "variables"),
	}, nil
}

// Create returns details for a new variable.
func (c *VariablesClient) Create(ctx context.Context, data api.VariableCreate) (*api.Variable, error) {
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.routePrefix+"/"+variableID.String(), http.NoBody)
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

	var variable api.Variable
	if err := json.NewDecoder(resp.Body).Decode(&variable); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &variable, nil
}

// GetByName returns details for a variable by name.
func (c *VariablesClient) GetByName(ctx context.Context, name string) (*api.Variable, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.routePrefix+"/name/"+name, http.NoBody)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.routePrefix+"/"+variableID.String(), &buf)
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

// Delete removes a variable by ID.
func (c *VariablesClient) Delete(ctx context.Context, variableID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.routePrefix+"/"+variableID.String(), http.NoBody)
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
