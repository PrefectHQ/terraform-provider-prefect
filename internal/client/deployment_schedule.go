package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.DeploymentScheduleClient(&DeploymentScheduleClient{})

type DeploymentScheduleClient struct {
	hc              *http.Client
	routePrefix     string
	apiKey          string
	basicAuthKey    string
	csrfClientToken string
	csrfToken       string
}

// DeploymentSchedule returns a DeploymentScheduleClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) DeploymentSchedule(accountID, workspaceID uuid.UUID) (api.DeploymentScheduleClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &DeploymentScheduleClient{
		hc:              c.hc,
		routePrefix:     getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "deployments"),
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}, nil
}

func (c *DeploymentScheduleClient) Create(ctx context.Context, deploymentID uuid.UUID, payload []api.DeploymentSchedulePayload) ([]*api.DeploymentSchedule, error) {
	cfg := requestConfig{
		method:          http.MethodPost,
		url:             fmt.Sprintf("%s/%s/schedules", c.routePrefix, deploymentID.String()),
		body:            &payload,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusCreated,
	}

	var schedules []*api.DeploymentSchedule
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &schedules); err != nil {
		return nil, fmt.Errorf("failed to create deployment schedules: %w", err)
	}

	return schedules, nil
}

func (c *DeploymentScheduleClient) Read(ctx context.Context, deploymentID uuid.UUID) ([]*api.DeploymentSchedule, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             fmt.Sprintf("%s/%s/schedules", c.routePrefix, deploymentID.String()),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var schedules []*api.DeploymentSchedule
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &schedules); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return schedules, nil
}

func (c *DeploymentScheduleClient) Update(ctx context.Context, deploymentID uuid.UUID, scheduleID uuid.UUID, payload api.DeploymentSchedulePayload) error {
	cfg := requestConfig{
		method:          http.MethodPatch,
		url:             fmt.Sprintf("%s/%s/%s/%s", c.routePrefix, deploymentID.String(), "schedules", scheduleID.String()),
		body:            &payload,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update deployment schedule: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (c *DeploymentScheduleClient) Delete(ctx context.Context, deploymentID uuid.UUID, scheduleID uuid.UUID) error {
	cfg := requestConfig{
		method:          http.MethodDelete,
		url:             fmt.Sprintf("%s/%s/%s/%s", c.routePrefix, deploymentID.String(), "schedules", scheduleID.String()),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete deployment schedule: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
