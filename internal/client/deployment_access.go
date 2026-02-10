package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.DeploymentAccessClient(&DeploymentAccessClient{})

type DeploymentAccessClient struct {
	hc              *http.Client
	routePrefix     string
	apiKey          string
	basicAuthKey    string
	csrfClientToken string
	csrfToken       string
	customHeaders   map[string]string
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

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &DeploymentAccessClient{
		hc:              c.hc,
		routePrefix:     getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "deployments"),
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
	}, nil
}

func (c *DeploymentAccessClient) Read(ctx context.Context, deploymentID uuid.UUID) (*api.DeploymentAccessControl, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             fmt.Sprintf("%s/%s/access", c.routePrefix, deploymentID.String()),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusOK,
	}

	var accessControl api.DeploymentAccessControl
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &accessControl); err != nil {
		return nil, fmt.Errorf("failed to get deployment access control: %w", err)
	}

	return &accessControl, nil
}

func (c *DeploymentAccessClient) Set(ctx context.Context, deploymentID uuid.UUID, accessControl api.DeploymentAccessSet) error {
	cfg := requestConfig{
		method:          http.MethodPut,
		url:             fmt.Sprintf("%s/%s/access", c.routePrefix, deploymentID.String()),
		body:            &accessControl,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to set deployment access control: %w", err)
	}

	defer resp.Body.Close()

	return nil
}
