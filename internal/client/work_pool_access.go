package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

type WorkPoolAccessClient struct {
	hc              *http.Client
	routePrefix     string
	apiKey          string
	basicAuthKey    string
	csrfClientToken string
	csrfToken       string
}

// WorkPoolAccess returns a WorkPoolAccessClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) WorkPoolAccess(accountID uuid.UUID, workspaceID uuid.UUID) (api.WorkPoolAccessClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &WorkPoolAccessClient{
		hc:              c.hc,
		routePrefix:     getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "work_pools"),
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}, nil
}

func (c *WorkPoolAccessClient) Read(ctx context.Context, workPoolName string) (*api.WorkPoolAccessControl, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             fmt.Sprintf("%s/%s/access", c.routePrefix, workPoolName),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var accessControl api.WorkPoolAccessControl
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &accessControl); err != nil {
		return nil, fmt.Errorf("failed to get work pool access control: %w", err)
	}

	return &accessControl, nil
}

func (c *WorkPoolAccessClient) Set(ctx context.Context, workPoolName string, accessControl api.WorkPoolAccessSet) error {
	cfg := requestConfig{
		method:          http.MethodPut,
		url:             fmt.Sprintf("%s/%s/access", c.routePrefix, workPoolName),
		body:            &accessControl,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to set work pool access control: %w", err)
	}

	defer resp.Body.Close()

	return nil
}
