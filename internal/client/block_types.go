package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = api.BlockTypeClient(&BlockTypeClient{})

// BlockTypeClient is a client for working with block types.
type BlockTypeClient struct {
	hc          *http.Client
	routePrefix string
	apiKey      string
}

// BlockTypes returns a BlockTypeClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) BlockTypes(accountID uuid.UUID, workspaceID uuid.UUID) (api.BlockTypeClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}
	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if helpers.IsCloudEndpoint(c.endpoint) && (accountID == uuid.Nil || workspaceID == uuid.Nil) {
		return nil, fmt.Errorf("prefect Cloud endpoints require an account_id and workspace_id to be set on either the provider or the resource")
	}

	return &BlockTypeClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "block_types"),
	}, nil
}

// GetBySlug returns details for a block type by slug.
func (c *BlockTypeClient) GetBySlug(ctx context.Context, slug string) (*api.BlockType, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.routePrefix+"/slug/"+slug, http.NoBody)
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

	var blockType api.BlockType
	if err := json.NewDecoder(resp.Body).Decode(&blockType); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &blockType, nil
}
