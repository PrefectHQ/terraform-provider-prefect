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
	// NOTE: our access APIs can optionally take a single "access" payload
	// or a list of them. In our case, we'll always pass a 1-item slice
	payloads := []api.WorkspaceAccessUpsert{
		{
			WorkspaceRoleID: roleID,
		},
	}
	var requestPath string

	// NOTE: this is a quirk of our <entity>_access API at the moment
	// where user_access and bot_access were originally set up as a POST.
	//
	// Semantically, they should be a PUT, which the newer team_access API is set up as.
	// At a later point, we will migrate the user/bot API variants over to a PUT
	// in a breaking change.
	var requestMethod string

	if accessorType == utils.User {
		requestPath = fmt.Sprintf("%s/user_access/", c.routePrefix)
		payloads[0].UserID = &accessorID
		requestMethod = http.MethodPost
	}
	if accessorType == utils.ServiceAccount {
		requestPath = fmt.Sprintf("%s/bot_access/", c.routePrefix)
		payloads[0].BotID = &accessorID
		requestMethod = http.MethodPost
	}
	if accessorType == utils.Team {
		requestPath = fmt.Sprintf("%s/team_access/", c.routePrefix)
		payloads[0].TeamID = &accessorID
		requestMethod = http.MethodPut
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&payloads); err != nil {
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

	var workspaceAccesses []api.WorkspaceAccess
	if err := json.NewDecoder(resp.Body).Decode(&workspaceAccesses); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &workspaceAccesses[0], nil
}

// Get fetches workspace access for various accessor types via accessID.
func (c *WorkspaceAccessClient) Get(ctx context.Context, accessorType string, accessID uuid.UUID) (*api.WorkspaceAccess, error) {
	var requestPath string
	var requestMethod string

	// NOTE: this is a quirk of our <entity>_access API at the moment
	// where user_access and bot_access can be fetched individually by resource ID,
	// whereas team_access resources must be fetched as a list, scoped to the workspace.
	//
	// Here, we'll key off of the `accessorType` to determine the correct API endpoint,
	// and we'll conditionally handle unmarshalling the response to
	// a single WorkspaceAccess (for user_access and bot_access)
	// or a list of them (for team_access) -> filtering for the passed `accessID`.
	if accessorType == utils.User {
		// GET: /.../<workspace_access_id>
		requestMethod = http.MethodGet
		requestPath = fmt.Sprintf("%s/user_access/%s", c.routePrefix, accessID.String())
	}
	if accessorType == utils.ServiceAccount {
		// GET: /.../<workspace_access_id>
		requestMethod = http.MethodGet
		requestPath = fmt.Sprintf("%s/bot_access/%s", c.routePrefix, accessID.String())
	}
	if accessorType == utils.Team {
		// POST: /.../filter
		requestMethod = http.MethodPost
		requestPath = fmt.Sprintf("%s/team_access/filter", c.routePrefix)
	}

	req, err := http.NewRequestWithContext(ctx, requestMethod, requestPath, http.NoBody)
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
	var workspaceAccesses []api.WorkspaceAccess

	// If this is a team_access resource, we'll expect a list of WorkspaceAccess objects
	if accessorType == utils.Team {
		if err := json.NewDecoder(resp.Body).Decode(&workspaceAccesses); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
	}
	for _, access := range workspaceAccesses {
		if access.ID == accessID {
			workspaceAccess = access

			break
		}
	}

	// Otherwise, we'll expect a single WorkspaceAccess object, fetched by `accessID`
	if accessorType == utils.User || accessorType == utils.ServiceAccount {
		if err := json.NewDecoder(resp.Body).Decode(&workspaceAccess); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
	}

	if workspaceAccess.ID == uuid.Nil {
		return nil, fmt.Errorf("workspace access not found for accessID: %s", accessID.String())
	}

	return &workspaceAccess, nil
}

// DeleteUserAccess deletes a service account's workspace access via accessID.
func (c *WorkspaceAccessClient) Delete(ctx context.Context, accessorType string, accessID uuid.UUID, accessorID uuid.UUID) error {
	var requestPath string

	// NOTE: this is a quirk of our <entity>_access API at the moment
	// where user_access and bot_access can be deleted individually by the `accessID`
	// whereas team_access resources must be deleted by the `accessorID`, eg. the team_id
	//
	// Here, we'll key off of the `accessorType` to determine the correct URL parameter to use.
	//
	// DELETE: /.../<workspace_access_id>
	if accessorType == utils.User {
		requestPath = fmt.Sprintf("%s/user_access/%s", c.routePrefix, accessID.String())
	}

	// DELETE: /.../<workspace_access_id>
	if accessorType == utils.ServiceAccount {
		requestPath = fmt.Sprintf("%s/bot_access/%s", c.routePrefix, accessID.String())
	}

	// DELETE: /.../<team_id>
	if accessorType == utils.Team {
		requestPath = fmt.Sprintf("%s/team_access/%s", c.routePrefix, accessorID.String())
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
