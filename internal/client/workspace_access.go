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

// type assertion ensures that this client implements the interface.
var _ = api.WorkspaceAccessClient(&WorkspaceAccessClient{})

type WorkspaceAccessClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}

// WorkspaceAccess is a factory that initializes and returns a WorkspaceAccessClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) WorkspaceAccess(accountID uuid.UUID, workspaceID uuid.UUID) (api.WorkspaceAccessClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}
	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}
	if accountID == uuid.Nil || workspaceID == uuid.Nil {
		return nil, fmt.Errorf("both accountID and workspaceID must be defined")
	}

	return &WorkspaceAccessClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: fmt.Sprintf("%s/accounts/%s/workspaces/%s", c.endpoint, accountID.String(), workspaceID.String()),
	}, nil
}

// UpsertServiceAccountAccess creates or updates a service account's access to a workspace.
func (c *WorkspaceAccessClient) UpsertServiceAccountAccess(ctx context.Context, payload api.WorkspaceAccessUpsert) (*api.WorkspaceServiceAccountAccess, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&payload); err != nil {
		return nil, fmt.Errorf("failed to encode create payload data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/bot_access/", c.routePrefix), &buf)
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
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var workspaceServiceAccountAccess api.WorkspaceServiceAccountAccess
	if err := json.NewDecoder(resp.Body).Decode(&workspaceServiceAccountAccess); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspaceServiceAccountAccess, nil
}

// UpsertUserAccess creates or updates a user's access to a workspace.
func (c *WorkspaceAccessClient) UpsertUserAccess(ctx context.Context, payload api.WorkspaceAccessUpsert) (*api.WorkspaceUserAccess, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&payload); err != nil {
		return nil, fmt.Errorf("failed to encode create payload data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/user_access/", c.routePrefix), &buf)
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
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var workspaceUserAccess api.WorkspaceUserAccess
	if err := json.NewDecoder(resp.Body).Decode(&workspaceUserAccess); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspaceUserAccess, nil
}

// GetServiceAccountAccess fetches a service account's workspace access via accessID.
func (c *WorkspaceAccessClient) GetServiceAccountAccess(ctx context.Context, accessID uuid.UUID) (*api.WorkspaceServiceAccountAccess, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/bot_access/%s", c.routePrefix, accessID.String()), http.NoBody)
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
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var workspaceServiceAccountAccess api.WorkspaceServiceAccountAccess
	if err := json.NewDecoder(resp.Body).Decode(&workspaceServiceAccountAccess); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspaceServiceAccountAccess, nil
}

// GetUserAccess fetches a user's workspace access via accessID.
func (c *WorkspaceAccessClient) GetUserAccess(ctx context.Context, accessID uuid.UUID) (*api.WorkspaceUserAccess, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/user_access/%s", c.routePrefix, accessID.String()), http.NoBody)
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
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var workspaceUserAccess api.WorkspaceUserAccess
	if err := json.NewDecoder(resp.Body).Decode(&workspaceUserAccess); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspaceUserAccess, nil
}

// DeleteServiceAccountAccess deletes a service account's workspace access via accessID.
func (c *WorkspaceAccessClient) DeleteServiceAccountAccess(ctx context.Context, accessID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/bot_access/%s", c.routePrefix, accessID.String()), http.NoBody)
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
		return fmt.Errorf("status code %s", resp.Status)
	}

	return nil
}

// DeleteUserAccess deletes a service account's workspace access via accessID.
func (c *WorkspaceAccessClient) DeleteUserAccess(ctx context.Context, accessID uuid.UUID) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/user_access/%s", c.routePrefix, accessID.String()), http.NoBody)
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
		return fmt.Errorf("status code %s", resp.Status)
	}

	return nil
}
