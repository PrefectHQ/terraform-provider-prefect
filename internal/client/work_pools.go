package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.WorkPoolsClient(&WorkPoolsClient{})

// WorkPoolsClient is a client for working with work pools.
type WorkPoolsClient struct {
	hc              *http.Client
	apiKey          string
	basicAuthKey    string
	routePrefix     string
	csrfClientToken string
	csrfToken       string
	customHeaders   map[string]string
}

// WorkPools returns a WorkPoolsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) WorkPools(accountID uuid.UUID, workspaceID uuid.UUID) (api.WorkPoolsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &WorkPoolsClient{
		hc:              c.hc,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		routePrefix:     getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "work_pools"),
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
	}, nil
}

// Create returns details for a new work pool.
func (c *WorkPoolsClient) Create(ctx context.Context, data api.WorkPoolCreate) (*api.WorkPool, error) {
	cfg := requestConfig{
		method:          http.MethodPost,
		url:             c.routePrefix + "/",
		body:            &data,
		successCodes:    successCodesStatusCreated,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
	}

	var pool api.WorkPool
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &pool); err != nil {
		return nil, fmt.Errorf("failed to create work pool: %w", err)
	}

	return &pool, nil
}

const workPoolsDefaultPageSize int64 = 200

// List returns a list of work pools matching filter criteria.
// It paginates through all results automatically using offset/limit.
func (c *WorkPoolsClient) List(ctx context.Context, ids []string) ([]*api.WorkPool, error) {
	var filterQuery api.WorkPoolFilterRequest
	if len(ids) > 0 {
		filterQuery.WorkPools.ID.Any = ids
	}

	var allPools []*api.WorkPool
	offset := int64(0)
	limit := workPoolsDefaultPageSize

	for {
		filterQuery.Offset = &offset
		filterQuery.Limit = &limit

		cfg := requestConfig{
			method:          http.MethodPost,
			url:             c.routePrefix + "/filter",
			body:            &filterQuery,
			successCodes:    successCodesStatusOK,
			apiKey:          c.apiKey,
			basicAuthKey:    c.basicAuthKey,
			csrfClientToken: c.csrfClientToken,
			csrfToken:       c.csrfToken,
			customHeaders:   c.customHeaders,
		}

		var page []*api.WorkPool
		if err := requestWithDecodeResponse(ctx, c.hc, cfg, &page); err != nil {
			return nil, fmt.Errorf("failed to list work pools: %w", err)
		}

		allPools = append(allPools, page...)

		if int64(len(page)) < limit {
			break
		}

		offset += limit
	}

	return allPools, nil
}

// Get returns details for a work pool by name.
func (c *WorkPoolsClient) Get(ctx context.Context, name string) (*api.WorkPool, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             c.routePrefix + "/" + name,
		successCodes:    successCodesStatusOK,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
	}

	var pool api.WorkPool
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &pool); err != nil {
		return nil, fmt.Errorf("failed to get work pool: %w", err)
	}

	return &pool, nil
}

// Update modifies an existing work pool by name.
func (c *WorkPoolsClient) Update(ctx context.Context, name string, data api.WorkPoolUpdate) error {
	cfg := requestConfig{
		method:          http.MethodPatch,
		url:             c.routePrefix + "/" + name,
		body:            &data,
		successCodes:    successCodesStatusOKOrNoContent,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update work pool: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Delete removes a work pool by name.
func (c *WorkPoolsClient) Delete(ctx context.Context, name string) error {
	cfg := requestConfig{
		method:          http.MethodDelete,
		url:             c.routePrefix + "/" + name,
		successCodes:    successCodesStatusOKOrNoContent,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete work pool: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
