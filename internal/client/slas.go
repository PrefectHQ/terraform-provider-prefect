package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.SLAsClient(&SLAsClient{})

// SLAsClient is a client for working with SLAs.
type SLAsClient struct {
	hc              *http.Client
	routePrefix     string
	apiKey          string
	basicAuthKey    string
	csrfClientToken string
	csrfToken       string
}

// SLAs returns a SLAsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) SLAs(accountID uuid.UUID, workspaceID uuid.UUID) (api.SLAsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &SLAsClient{
		hc:              c.hc,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		routePrefix:     getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "slas"),
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}, nil
}

// ApplyResourceSLAs applies SLAs to a resource.
func (c *SLAsClient) ApplyResourceSLAs(ctx context.Context, resourceID string, slas []api.SLAUpsert) (*api.SLAResponse, error) {
	cfg := requestConfig{
		method:          http.MethodPost,
		url:             fmt.Sprintf("%s/apply-resource-slas/%s", c.routePrefix, resourceID),
		body:            slas,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var response api.SLAResponse
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &response); err != nil {
		return nil, err
	}

	return &response, nil
}
