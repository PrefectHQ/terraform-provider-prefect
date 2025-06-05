package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

// BlockSchemaClient is a client for working with block schemas.
type BlockSchemaClient struct {
	hc           *http.Client
	routePrefix  string
	apiKey       string
	basicAuthKey string
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

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &BlockSchemaClient{
		hc:           c.hc,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		routePrefix:  getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "block_schemas"),
	}, nil
}

// List gets a list of BlockSchemas for a given list of block type slugs.
func (c *BlockSchemaClient) List(ctx context.Context, blockTypeIDs []uuid.UUID) ([]*api.BlockSchema, error) {
	filterQuery := &api.BlockSchemaFilter{}
	filterQuery.BlockSchemas.BlockTypeID.Any = blockTypeIDs

	cfg := requestConfig{
		method:       http.MethodPost,
		url:          c.routePrefix + "/filter",
		body:         filterQuery,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusOK,
	}

	var blockSchemas []*api.BlockSchema
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &blockSchemas); err != nil {
		return nil, fmt.Errorf("failed to get block schemas: %w", err)
	}

	return blockSchemas, nil
}

// Create creates a new BlockSchema.
func (c *BlockSchemaClient) Create(ctx context.Context, payload *api.BlockSchemaCreate) (*api.BlockSchema, error) {
	cfg := requestConfig{
		method:       http.MethodPost,
		url:          c.routePrefix,
		body:         payload,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusCreated,
	}

	var createdBlockSchema *api.BlockSchema
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &createdBlockSchema); err != nil {
		return nil, fmt.Errorf("failed to create block schema: %w", err)
	}

	return createdBlockSchema, nil
}

// Read gets a BlockSchema by ID.
func (c *BlockSchemaClient) Read(ctx context.Context, id uuid.UUID) (*api.BlockSchema, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          c.routePrefix + "/" + id.String(),
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusOK,
	}

	var blockSchema *api.BlockSchema
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &blockSchema); err != nil {
		return nil, fmt.Errorf("failed to get block schema: %w", err)
	}

	return blockSchema, nil
}

// Delete deletes a BlockSchema by ID.
func (c *BlockSchemaClient) Delete(ctx context.Context, id uuid.UUID) error {
	cfg := requestConfig{
		method:       http.MethodDelete,
		url:          c.routePrefix + "/" + id.String(),
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete block schema: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
