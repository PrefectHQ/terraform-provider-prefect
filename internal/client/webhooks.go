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

type webhooksClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}

func (c *Client) Webhooks(accountID, workspaceID uuid.UUID) (api.WebhooksClient, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("apiKey is not set")
	}

	if c.endpoint == "" {
		return nil, fmt.Errorf("endpoint is not set")
	}

	routePrefix := fmt.Sprintf("%s/api/accounts/%s/workspaces/%s/webhooks", c.endpoint, accountID, workspaceID)

	return &webhooksClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: routePrefix,
	}, nil
}

func (wc *webhooksClient) Create(ctx context.Context, accountID, workspaceID string, request api.WebhookCreateRequest) (*api.Webhook, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&request); err != nil {
		return nil, fmt.Errorf("failed to encode request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, wc.routePrefix+"/", &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, wc.apiKey)

	resp, err := wc.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	var response api.Webhook
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (wc *webhooksClient) Get(ctx context.Context, accountID, workspaceID, webhookID string) (*api.Webhook, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wc.routePrefix+"/"+webhookID, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, wc.apiKey)

	resp, err := wc.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	var response api.Webhook
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (wc *webhooksClient) Update(ctx context.Context, accountID, workspaceID, webhookID string, request api.WebhookUpdateRequest) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&request); err != nil {
		return fmt.Errorf("failed to encode request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, wc.routePrefix+"/"+webhookID, &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, wc.apiKey)

	resp, err := wc.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	return nil
}

func (wc *webhooksClient) Delete(ctx context.Context, accountID, workspaceID, webhookID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, wc.routePrefix+"/"+webhookID, http.NoBody)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, wc.apiKey)

	resp, err := wc.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	return nil
}

func (w *webhooksClient) List(ctx context.Context, accountID, workspaceID string) ([]*api.Webhook, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/", w.routePrefix), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, w.apiKey)

	resp, err := w.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	var webhooks []*api.Webhook
	if err := json.NewDecoder(resp.Body).Decode(&webhooks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return webhooks, nil
}