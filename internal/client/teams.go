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

// Create creates a team.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *TeamsClient) Create(ctx context.Context, payload api.TeamCreate) (*api.Team, error) {
	cfg := requestConfig{
		method:       http.MethodPost,
		url:          c.routePrefix,
		body:         payload,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusCreated,
	}

	var team api.Team
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &team); err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return &team, nil
}

// Read reads a team.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *TeamsClient) Read(ctx context.Context, teamID string) (*api.Team, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, teamID),
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusOK,
	}

	var team api.Team
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &team); err != nil {
		return nil, fmt.Errorf("failed to read team: %w", err)
	}

	return &team, nil
}

// Update updates a team.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *TeamsClient) Update(ctx context.Context, teamID string, payload api.TeamUpdate) (*api.Team, error) {
	cfg := requestConfig{
		method:       http.MethodPut,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, teamID),
		body:         payload,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusOK,
	}

	var team api.Team
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &team); err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	return &team, nil
}

// Delete deletes a team.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *TeamsClient) Delete(ctx context.Context, teamID string) error {
	cfg := requestConfig{
		method:       http.MethodDelete,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, teamID),
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	defer resp.Body.Close()

	return nil
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
