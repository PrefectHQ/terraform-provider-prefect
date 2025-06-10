package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.DeploymentsClient(&DeploymentsClient{})

// DeploymentsClient is a client for working with Deployments.
type DeploymentsClient struct {
	hc              *http.Client
	routePrefix     string
	apiKey          string
	basicAuthKey    string
	csrfClientToken string
	csrfToken       string
}

// Deployments returns a DeploymentsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Deployments(accountID uuid.UUID, workspaceID uuid.UUID) (api.DeploymentsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &DeploymentsClient{
		hc:              c.hc,
		routePrefix:     getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "deployments"),
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}, nil
}

// Create returns details for a new Deployment.
func (c *DeploymentsClient) Create(ctx context.Context, data api.DeploymentCreate) (*api.Deployment, error) {
	cfg := requestConfig{
		method:          http.MethodPost,
		url:             c.routePrefix + "/",
		body:            &data,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusCreated,
	}

	var deployment api.Deployment
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &deployment); err != nil {
		return nil, fmt.Errorf("failed to create deployment: %w", err)
	}

	return &deployment, nil
}

// Get returns details for a Deployment by ID.
func (c *DeploymentsClient) Get(ctx context.Context, deploymentID uuid.UUID) (*api.Deployment, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, deploymentID.String()),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var deployment api.Deployment
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &deployment); err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return &deployment, nil
}

// GetByName returns details for a Deployment by name.
func (c *DeploymentsClient) GetByName(ctx context.Context, flowName, deploymentName string) (*api.Deployment, error) {
	url := fmt.Sprintf("%s/name/%s/%s", c.routePrefix, flowName, deploymentName)
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             url,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var deployment api.Deployment
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &deployment); err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	return &deployment, nil
}

// Update modifies an existing Deployment by ID.
func (c *DeploymentsClient) Update(ctx context.Context, id uuid.UUID, data api.DeploymentUpdate) error {
	cfg := requestConfig{
		method:          http.MethodPatch,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, id.String()),
		body:            &data,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Delete removes a Deployment by ID.
func (c *DeploymentsClient) Delete(ctx context.Context, deploymentID uuid.UUID) error {
	cfg := requestConfig{
		method:          http.MethodDelete,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, deploymentID.String()),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete deployment: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
