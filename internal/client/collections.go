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

var _ = api.CollectionsClient(&CollectionsClient{})

type CollectionsClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}

// Collections returns an CollectionsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Collections(accountID, workspaceID uuid.UUID) (api.CollectionsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if helpers.IsCloudEndpoint(c.endpoint) && (accountID == uuid.Nil || workspaceID == uuid.Nil) {
		return nil, fmt.Errorf("prefect Cloud endpoints require an account_id and workspace_id to be set on either the provider or the resource")
	}

	return &CollectionsClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "collections"),
	}, nil
}

// GetWorkerMetadataViews returns a map of worker metadata views by prefect package name.
// This endpoint serves base job configurations for the primary worker types.
func (c *CollectionsClient) GetWorkerMetadataViews(ctx context.Context) (api.WorkerTypeByPackage, error) {
	routeSuffix := "views/aggregate-worker-metadata"
	if helpers.IsCloudEndpoint(c.routePrefix) {
		routeSuffix = "work_pool_types"
	}

	url := fmt.Sprintf("%s/%s", c.routePrefix, routeSuffix)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
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
