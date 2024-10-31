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

var _ = api.DeploymentAccessClient(&DeploymentAccessClient{})

type DeploymentAccessClient struct {
	hc          *http.Client
	routePrefix string
	apiKey      string
}

// DeploymentAccess returns a DeploymentAccessClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) DeploymentAccess(accountID uuid.UUID, workspaceID uuid.UUID) (api.DeploymentAccessClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	return &DeploymentAccessClient{
		hc:          c.hc,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "deployments"),
		apiKey:      c.apiKey,
	}, nil
}

func (c *DeploymentAccessClient) Read(ctx context.Context, deploymentID uuid.UUID) (*api.DeploymentAccessControl, error) {
	url := fmt.Sprintf("%s/%s/access", c.routePrefix, deploymentID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
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

	var accessControl api.DeploymentAccessControl
	if err := json.NewDecoder(resp.Body).Decode(&accessControl); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &accessControl, nil
}

func (c *DeploymentAccessClient) Set(ctx context.Context, deploymentID uuid.UUID, accessControl api.DeploymentAccessSet) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&accessControl); err != nil {
		return fmt.Errorf("failed to encode access control: %w", err)
	}

	url := fmt.Sprintf("%s/%s/access", c.routePrefix, deploymentID.String())

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, &buf)
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
