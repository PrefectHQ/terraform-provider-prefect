package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/provider/helpers"
)

var _ = api.CollectionsClient(&CollectionsClient{})

type CollectionsClient struct {
	hc           *http.Client
	apiKey       string
	basicAuthKey string
	routePrefix  string
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

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &CollectionsClient{
		hc:           c.hc,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		routePrefix:  getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "collections"),
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

	cfg := requestConfig{
		method:       http.MethodGet,
		url:          url,
		body:         http.NoBody,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusOK,
	}

	var workerTypeByPackage api.WorkerTypeByPackage
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &workerTypeByPackage); err != nil {
		return nil, fmt.Errorf("failed to get worker type by package: %w", err)
	}

	return workerTypeByPackage, nil
}
