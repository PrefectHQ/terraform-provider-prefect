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
	"github.com/prefecthq/terraform-provider-prefect/internal/utils"
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

// Upsert creates or updates access to a workspace for various accessor types.
func (c *WorkspaceAccessClient) Upsert(ctx context.Context, accessorType string, accessorID uuid.UUID, roleID uuid.UUID) (*api.WorkspaceAccess, error) {
	payload := api.WorkspaceAccessUpsert{
		WorkspaceRoleID: roleID,
	}
	var requestPath string

	// NOTE: this is a quirk of our <entity>_access API at the moment
	// where user_access and bot_access were originally set up as a POST.
	// Semantically, they should be a PUT, which the newer team_access API is set up as.
	// At a later point, we will migrate the user/bot API variants over to a PUT
	// in a breaking change.
	requestMethod := http.MethodPut

	if accessorType == utils.User {
		requestPath = fmt.Sprintf("%s/user_access/", c.routePrefix)
		payload.UserID = &accessorID
		requestMethod = http.MethodPost
	}
	if accessorType == utils.ServiceAccount {
		requestPath = fmt.Sprintf("%s/bot_access/", c.routePrefix)
		payload.BotID = &accessorID
		requestMethod = http.MethodPost
	}
	if accessorType == utils.Team {
		requestPath = fmt.Sprintf("%s/team_access/", c.routePrefix)
		payload.TeamID = &accessorID
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&payload); err != nil {
		return nil, fmt.Errorf("failed to encode upsert payload data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, requestMethod, requestPath, &buf)
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

	var workspaceAccess api.WorkspaceAccess
	if err := json.NewDecoder(resp.Body).Decode(&workspaceAccess); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspaceAccess, nil
}

// Get fetches workspace access for various accessor types via accessID.
func (c *WorkspaceAccessClient) Get(ctx context.Context, accessorType string, accessID uuid.UUID) (*api.WorkspaceAccess, error) {
	var requestPath string
	if accessorType == utils.User {
		requestPath = fmt.Sprintf("%s/user_access/%s", c.routePrefix, accessID.String())
	}
	if accessorType == utils.ServiceAccount {
		requestPath = fmt.Sprintf("%s/bot_access/%s", c.routePrefix, accessID.String())
	}
	if accessorType == utils.Team {
		requestPath = fmt.Sprintf("%s/team_access/%s", c.routePrefix, accessID.String())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestPath, http.NoBody)
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

	var workspaceAccess api.WorkspaceAccess
	if err := json.NewDecoder(resp.Body).Decode(&workspaceAccess); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspaceAccess, nil
}

// DeleteUserAccess deletes a service account's workspace access via accessID.
func (c *WorkspaceAccessClient) Delete(ctx context.Context, accessorType string, accessID uuid.UUID) error {
	var requestPath string
	if accessorType == utils.User {
		requestPath = fmt.Sprintf("%s/user_access/%s", c.routePrefix, accessID.String())
	}
	if accessorType == utils.ServiceAccount {
		requestPath = fmt.Sprintf("%s/bot_access/%s", c.routePrefix, accessID.String())
	}
	if accessorType == utils.Team {
		requestPath = fmt.Sprintf("%s/team_access/%s", c.routePrefix, accessID.String())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, requestPath, http.NoBody)
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
