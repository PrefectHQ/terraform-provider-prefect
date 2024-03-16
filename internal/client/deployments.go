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

var _ = api.DeploymentsClient(&DeploymentsClient{})

// DeploymentsClient is a client for working with Deployments.
type DeploymentsClient struct {
	hc          *http.Client
	routePrefix string
	apiKey      string
}

// Deployments returns a DeploymentsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Deployments(accountID uuid.UUID, workspaceID uuid.UUID) (api.DeploymentsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	return &DeploymentsClient{
		hc:          c.hc,
		routePrefix: getAccountScopedURL(c.endpoint, accountID, "workspaces"),
		apiKey:      c.apiKey,
	}, nil
}

// Create returns details for a new Workspace.
func (c *DeploymentsClient) Create(ctx context.Context, data api.DeploymentCreate) (*api.Deployment, error) {
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

	var deployment api.Deployment
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &deployment, nil
}

// List returns a list of Workspaces, based on the provided list of handle names.
func (c *DeploymentsClient) List(ctx context.Context, handleNames []string) ([]*api.Deployment, error) {
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

	var deployments []*api.Deployment
	if err := json.NewDecoder(resp.Body).Decode(&deployments); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return deployments, nil
}

// Get returns details for a Workspace by ID.
func (c *DeploymentsClient) Get(ctx context.Context, workspaceID uuid.UUID) (*api.Deployment, error) {
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

	var deployment api.Deployment
	if err := json.NewDecoder(resp.Body).Decode(&deployment); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &deployment, nil
}

// Update modifies an existing Workspace by ID.
func (c *DeploymentsClient) Update(ctx context.Context, workspaceID uuid.UUID, data api.DeploymentUpdate) error {
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

// Delete removes a Deployment by ID.
func (c *DeploymentsClient) Delete(ctx context.Context, deploymentID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.routePrefix+"/"+deploymentID.String(), http.NoBody)
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
