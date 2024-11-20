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
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = api.DeploymentScheduleClient(&DeploymentScheduleClient{})

type DeploymentScheduleClient struct {
	hc          *http.Client
	routePrefix string
	apiKey      string
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

	if helpers.IsCloudEndpoint(c.endpoint) && (accountID == uuid.Nil || workspaceID == uuid.Nil) {
		return nil, fmt.Errorf("prefect Cloud endpoints require an account_id and workspace_id to be set on either the provider or the resource")
	}

	return &DeploymentScheduleClient{
		hc:          c.hc,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "deployments"),
		apiKey:      c.apiKey,
	}, nil
}

func (c *DeploymentScheduleClient) Create(ctx context.Context, deploymentID uuid.UUID, payload []api.DeploymentSchedulePayload) ([]*api.DeploymentSchedule, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(payload); err != nil {
		return nil, fmt.Errorf("error encoding payload: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s", c.routePrefix, deploymentID, "schedules")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &buf)
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

	var schedules []*api.DeploymentSchedule
	if err := json.NewDecoder(resp.Body).Decode(&schedules); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return schedules, nil
}

func (c *DeploymentScheduleClient) Read(ctx context.Context, deploymentID uuid.UUID) ([]*api.DeploymentSchedule, error) {
	url := fmt.Sprintf("%s/%s/%s", c.routePrefix, deploymentID, "schedules")
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

	var schedules []*api.DeploymentSchedule
	if err := json.NewDecoder(resp.Body).Decode(&schedules); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return schedules, nil
}

func (c *DeploymentScheduleClient) Update(ctx context.Context, deploymentID uuid.UUID, scheduleID uuid.UUID, payload api.DeploymentSchedulePayload) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&payload); err != nil {
		return fmt.Errorf("failed to encode update payload data: %w", err)
	}

	url := fmt.Sprintf("%s/%s/%s/%s", c.routePrefix, deploymentID.String(), "schedules", scheduleID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, &buf)
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

// todo: delete pointed at 0000 uuid
func (c *DeploymentScheduleClient) Delete(ctx context.Context, deploymentID uuid.UUID, scheduleID uuid.UUID) error {
	url := fmt.Sprintf("%s/%s/%s/%s", c.routePrefix, deploymentID.String(), "schedules", scheduleID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, http.NoBody)
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
