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
)

var _ = api.BlockDocumentClient(&BlockDocumentClient{})

type BlockDocumentClient struct {
	hc          *http.Client
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
	if accountID == uuid.Nil && workspaceID == uuid.Nil {
		return nil, fmt.Errorf("account id or workspace id is required")
	}

	return &BlockDocumentClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: fmt.Sprintf("/account/%s/workspace/%s/block_documents", accountID.String(), workspaceID.String()),
	}, nil
}

func (c *BlockDocumentClient) Get(ctx context.Context, id uuid.UUID) (*api.BlockDocument, error) {
	reqURL := fmt.Sprintf("%s/%s", c.routePrefix, id.String())
	reqURL = fmt.Sprintf("%s?include_secrets=true", reqURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/", c.routePrefix), &buf)
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

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", c.routePrefix, id.String()), &buf)
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
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", c.routePrefix, id.String()), http.NoBody)
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

func (c *BlockDocumentClient) GetACL(ctx context.Context, id uuid.UUID) (*api.BlockDocumentAccess, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s/access", c.routePrefix, id.String()), http.NoBody)
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

func (c *BlockDocumentClient) UpdateACL(ctx context.Context, id uuid.UUID, payload api.
	BlockDocumentAccessReplace) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&payload); err != nil {
		return fmt.Errorf("failed to encode update payload data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("%s/%s/access", c.routePrefix, id.String()), &buf)
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