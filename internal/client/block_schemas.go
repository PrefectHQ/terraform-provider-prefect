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

// BlockSchemaClient is a client for working with block schemas.
type BlockSchemaClient struct {
	hc          *http.Client
	routePrefix string
	apiKey      string
}

// BlockSchemas returns a BlockSchemaClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) BlockSchemas(accountID uuid.UUID, workspaceID uuid.UUID) (api.BlockSchemaClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}
	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if helpers.IsCloudEndpoint(c.endpoint) && (accountID == uuid.Nil || workspaceID == uuid.Nil) {
		return nil, fmt.Errorf("prefect Cloud endpoints require an account_id and workspace_id to be set on either the provider or the resource")
	}

	return &BlockSchemaClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "block_schemas"),
	}, nil
}

// List gets a list of BlockSchemas for a given list of block type slugs.
func (c *BlockSchemaClient) List(ctx context.Context, blockTypeIDs []uuid.UUID) ([]*api.BlockSchema, error) {
	filterQuery := &api.BlockSchemaFilter{}
	filterQuery.BlockSchemas.BlockTypeID.Any = blockTypeIDs

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&filterQuery); err != nil {
		return nil, fmt.Errorf("failed to encode create payload data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.routePrefix+"/filter", &buf)
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

	var blockSchemas []*api.BlockSchema
	if err := json.NewDecoder(resp.Body).Decode(&blockSchemas); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return blockSchemas, nil
}
