package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.TeamsClient(&TeamsClient{})

type TeamsClient struct {
	hc           *http.Client
	apiKey       string
	basicAuthKey string
	routePrefix  string
}

// Teams is a factory that initializes and returns a TeamsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Teams(accountID uuid.UUID) (api.TeamsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	return &TeamsClient{
		hc:           c.hc,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		routePrefix:  getAccountScopedURL(c.endpoint, accountID, "teams"),
	}, nil
}

// List returns a list of teams, based on the provided filter.
func (c *TeamsClient) List(ctx context.Context, names []string) ([]*api.Team, error) {
	filterQuery := api.TeamFilter{}
	filterQuery.Teams.Name.Any = names

	cfg := requestConfig{
		method:       http.MethodPost,
		url:          fmt.Sprintf("%s/filter", c.routePrefix),
		body:         &filterQuery,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusOK,
	}

	var teams []*api.Team
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &teams); err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}

	return teams, nil
}
