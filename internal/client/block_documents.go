package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = api.BlockDocumentClient(&BlockDocumentClient{})

type BlockDocumentClient struct {
	hc          *retryablehttp.Client
	apiKey      string
	routePrefix string
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

	if helpers.IsCloudEndpoint(c.endpoint) && (accountID == uuid.Nil || workspaceID == uuid.Nil) {
		return nil, fmt.Errorf("prefect Cloud endpoints require an account_id and workspace_id to be set on either the provider or the resource")
	}

	return &BlockDocumentClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "block_documents"),
	}, nil
}

func (c *BlockDocumentClient) Get(ctx context.Context, id uuid.UUID) (*api.BlockDocument, error) {
	reqURL := fmt.Sprintf("%s/%s", c.routePrefix, id.String())
	reqURL = fmt.Sprintf("%s?include_secrets=true", reqURL)

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
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

	var blockDocument api.BlockDocument
	if err := json.NewDecoder(resp.Body).Decode(&blockDocument); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &blockDocument, nil
}

func (c *BlockDocumentClient) GetByName(ctx context.Context, typeSlug, name string) (*api.BlockDocument, error) {
	// This URL is a little different, as it starts with 'block_types' instead of 'block_documents'.
	newRoutePrefix := fmt.Sprintf("block_types/slug/%s/block_documents/name/%s", typeSlug, name)
	reqURL := strings.ReplaceAll(c.routePrefix, "block_documents", newRoutePrefix)
	reqURL = fmt.Sprintf("%s?include_secrets=true", reqURL)

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
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

	var blockDocument api.BlockDocument
	if err := json.NewDecoder(resp.Body).Decode(&blockDocument); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &blockDocument, nil
}

func (c *BlockDocumentClient) Create(ctx context.Context, payload api.BlockDocumentCreate) (*api.BlockDocument, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&payload); err != nil {
		return nil, fmt.Errorf("failed to encode create payload data: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, c.routePrefix+"/", &buf)
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

	var blockDocument api.BlockDocument
	if err := json.NewDecoder(resp.Body).Decode(&blockDocument); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &blockDocument, nil
}

func (c *BlockDocumentClient) Update(ctx context.Context, id uuid.UUID, payload api.BlockDocumentUpdate) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&payload); err != nil {
		return fmt.Errorf("failed to encode update payload data: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", c.routePrefix, id.String()), &buf)
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

func (c *BlockDocumentClient) Delete(ctx context.Context, id uuid.UUID) error {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", c.routePrefix, id.String()), http.NoBody)
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

func (c *BlockDocumentClient) GetAccess(ctx context.Context, id uuid.UUID) (*api.BlockDocumentAccess, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s/access", c.routePrefix, id.String()), http.NoBody)
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

	var blockDocumentAccess api.BlockDocumentAccess
	if err := json.NewDecoder(resp.Body).Decode(&blockDocumentAccess); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &blockDocumentAccess, nil
}

func (c *BlockDocumentClient) UpsertAccess(ctx context.Context, id uuid.UUID, payload api.BlockDocumentAccessUpsert) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&payload); err != nil {
		return fmt.Errorf("failed to encode update payload data: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/%s/access", c.routePrefix, id.String()), &buf)
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
