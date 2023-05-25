package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.WorkPoolsClient(&WorkPoolsClient{})

// WorkPoolsClient is a client for working with work pools.
type WorkPoolsClient struct {
	hc       *http.Client
	endpoint string
	apiKey   string
}

// WorkPools returns a WorkPoolsClient.
//
//nolint:ireturn
func (c *Client) WorkPools() api.WorkPoolsClient {
	return &WorkPoolsClient{
		hc:       c.hc,
		endpoint: c.endpoint,
		apiKey:   c.apiKey,
	}
}

func (c *WorkPoolsClient) Get(ctx context.Context, name string) (*api.WorkPool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.endpoint+"/work_pools/"+name, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := doRequest(c.hc, c.apiKey, req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var pool api.WorkPool
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pool, nil
}

func (c *WorkPoolsClient) Create(ctx context.Context, data api.WorkPoolCreate) (*api.WorkPool, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return nil, fmt.Errorf("failed to encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/work_pools", &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	resp, err := doRequest(c.hc, c.apiKey, req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var pool api.WorkPool
	if err := json.NewDecoder(resp.Body).Decode(&pool); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pool, nil
}

func (c *WorkPoolsClient) Update(_ context.Context, _ api.WorkPoolUpdate) error {
	return nil
}

func (c *WorkPoolsClient) Delete(_ context.Context, _ string) error {
	return nil
}
