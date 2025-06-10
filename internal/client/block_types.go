package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.BlockTypeClient(&BlockTypeClient{})

// BlockTypeClient is a client for working with block types.
type BlockTypeClient struct {
	hc              *http.Client
	routePrefix     string
	apiKey          string
	basicAuthKey    string
	csrfClientToken string
	csrfToken       string
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

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &BlockTypeClient{
		hc:              c.hc,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		routePrefix:     getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "block_types"),
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}, nil
}

// GetBySlug returns details for a block type by slug.
func (c *BlockTypeClient) GetBySlug(ctx context.Context, slug string) (*api.BlockType, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             c.routePrefix + "/slug/" + slug,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var blockType api.BlockType
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &blockType); err != nil {
		return nil, fmt.Errorf("failed to get block type: %w", err)
	}

	return &blockType, nil
}

// Create creates a new BlockType.
func (c *BlockTypeClient) Create(ctx context.Context, payload *api.BlockTypeCreate) (*api.BlockType, error) {
	cfg := requestConfig{
		method:       http.MethodPost,
		url:          c.routePrefix,
		body:         payload,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusCreated,
	}

	var createdBlockType *api.BlockType
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &createdBlockType); err != nil {
		return nil, fmt.Errorf("failed to create block type: %w", err)
	}

	return createdBlockType, nil
}

// Update updates a BlockType.
func (c *BlockTypeClient) Update(ctx context.Context, id uuid.UUID, payload *api.BlockTypeUpdate) error {
	cfg := requestConfig{
		method:       http.MethodPatch,
		url:          c.routePrefix + "/" + id.String(),
		body:         payload,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update block type: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Delete deletes a BlockType.
func (c *BlockTypeClient) Delete(ctx context.Context, id uuid.UUID) error {
	cfg := requestConfig{
		method:       http.MethodDelete,
		url:          c.routePrefix + "/" + id.String(),
		body:         http.NoBody,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete block type: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
