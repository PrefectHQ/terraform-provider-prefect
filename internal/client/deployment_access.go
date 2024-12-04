package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
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

	if helpers.IsCloudEndpoint(c.endpoint) && (accountID == uuid.Nil || workspaceID == uuid.Nil) {
		return nil, fmt.Errorf("prefect Cloud endpoints require an account_id and workspace_id to be set on either the provider or the resource")
	}

	return &DeploymentAccessClient{
		hc:          c.hc,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "deployments"),
		apiKey:      c.apiKey,
	}, nil
}

func (c *DeploymentAccessClient) Read(ctx context.Context, deploymentID uuid.UUID) (*api.DeploymentAccessControl, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          fmt.Sprintf("%s/%s/access", c.routePrefix, deploymentID.String()),
		body:         http.NoBody,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOK,
	}

	var accessControl api.DeploymentAccessControl
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &accessControl); err != nil {
		return nil, fmt.Errorf("failed to get deployment access control: %w", err)
	}

	return &accessControl, nil
}

func (c *DeploymentAccessClient) Set(ctx context.Context, deploymentID uuid.UUID, accessControl api.DeploymentAccessSet) error {
	cfg := requestConfig{
		method:       http.MethodPut,
		url:          fmt.Sprintf("%s/%s/access", c.routePrefix, deploymentID.String()),
		body:         &accessControl,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to set deployment access control: %w", err)
	}

	defer resp.Body.Close()

	return nil
}
