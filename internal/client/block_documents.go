package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.BlockDocumentClient(&BlockDocumentClient{})

type BlockDocumentClient struct {
	hc              *http.Client
	apiKey          string
	basicAuthKey    string
	routePrefix     string
	csrfClientToken string
	csrfToken       string
}

// BlockDocuments is a factory that initializes and returns a BlockDocumentClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) BlockDocuments(accountID uuid.UUID, workspaceID uuid.UUID) (api.BlockDocumentClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &BlockDocumentClient{
		hc:              c.hc,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		routePrefix:     getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "block_documents"),
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}, nil
}

func (c *BlockDocumentClient) Get(ctx context.Context, id uuid.UUID) (*api.BlockDocument, error) {
	reqURL := fmt.Sprintf("%s/%s", c.routePrefix, id.String())
	reqURL = fmt.Sprintf("%s?include_secrets=true", reqURL)

	cfg := requestConfig{
		method:          http.MethodGet,
		url:             reqURL,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var blockDocument api.BlockDocument
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &blockDocument); err != nil {
		return nil, fmt.Errorf("failed to get block document: %w", err)
	}

	return &blockDocument, nil
}

func (c *BlockDocumentClient) GetByName(ctx context.Context, typeSlug, name string) (*api.BlockDocument, error) {
	// This URL is a little different, as it starts with 'block_types' instead of 'block_documents'.
	newRoutePrefix := fmt.Sprintf("block_types/slug/%s/block_documents/name/%s", typeSlug, name)
	reqURL := strings.ReplaceAll(c.routePrefix, "block_documents", newRoutePrefix)
	reqURL = fmt.Sprintf("%s?include_secrets=true", reqURL)

	cfg := requestConfig{
		method:          http.MethodGet,
		url:             reqURL,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var blockDocument api.BlockDocument
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &blockDocument); err != nil {
		return nil, fmt.Errorf("failed to get block document: %w", err)
	}

	return &blockDocument, nil
}

func (c *BlockDocumentClient) Create(ctx context.Context, payload api.BlockDocumentCreate) (*api.BlockDocument, error) {
	cfg := requestConfig{
		method:          http.MethodPost,
		url:             c.routePrefix + "/",
		body:            payload,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusCreated,
	}

	var blockDocument api.BlockDocument
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &blockDocument); err != nil {
		return nil, fmt.Errorf("failed to create block document: %w", err)
	}

	return &blockDocument, nil
}

func (c *BlockDocumentClient) Update(ctx context.Context, id uuid.UUID, payload api.BlockDocumentUpdate) error {
	cfg := requestConfig{
		method:          http.MethodPatch,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, id.String()),
		body:            payload,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update block document: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (c *BlockDocumentClient) Delete(ctx context.Context, id uuid.UUID) error {
	cfg := requestConfig{
		method:          http.MethodDelete,
		url:             fmt.Sprintf("%s/%s", c.routePrefix, id.String()),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete block document: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

func (c *BlockDocumentClient) GetAccess(ctx context.Context, id uuid.UUID) (*api.BlockDocumentAccess, error) {
	reqURL := fmt.Sprintf("%s/%s/access", c.routePrefix, id.String())

	cfg := requestConfig{
		method:          http.MethodGet,
		url:             reqURL,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var blockDocumentAccess api.BlockDocumentAccess
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &blockDocumentAccess); err != nil {
		return nil, fmt.Errorf("failed to get block document access: %w", err)
	}

	return &blockDocumentAccess, nil
}

func (c *BlockDocumentClient) UpsertAccess(ctx context.Context, id uuid.UUID, payload api.BlockDocumentAccessUpsert) error {
	cfg := requestConfig{
		method:          http.MethodPut,
		url:             fmt.Sprintf("%s/%s/access", c.routePrefix, id.String()),
		body:            payload,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to upsert block document access: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
