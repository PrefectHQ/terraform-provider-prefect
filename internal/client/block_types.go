package client

import (
	"context"
	"fmt"
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
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          c.routePrefix + "/slug/" + slug,
		body:         http.NoBody,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOK,
	}

	var blockType api.BlockType
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &blockType); err != nil {
		return nil, fmt.Errorf("failed to get block type: %w", err)
	}

	return &blockType, nil
}
