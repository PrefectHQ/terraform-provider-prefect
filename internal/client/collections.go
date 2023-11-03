package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.CollectionsClient(&CollectionsClient{})

type CollectionsClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}

// Collections returns an CollectionsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Collections() (api.CollectionsClient, error) {
	return &CollectionsClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: fmt.Sprintf("%s/collections", c.endpoint),
	}, nil
}

// GetWorkerMetadataViews returns a map of worker metadata views by prefect package name.
// This endpoint serves base job configurations for the primary worker types.
func (c *CollectionsClient) GetWorkerMetadataViews(ctx context.Context) (api.WorkerTypeByPackage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/views/aggregate-worker-metadata", c.routePrefix), http.NoBody)
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

	var workerTypeByPackage api.WorkerTypeByPackage
	if err := json.NewDecoder(resp.Body).Decode(&workerTypeByPackage); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return workerTypeByPackage, nil
}
