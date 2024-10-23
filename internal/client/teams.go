package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.TeamsClient(&TeamsClient{})

type TeamsClient struct {
	hc          *retryablehttp.Client
	apiKey      string
	routePrefix string
}

// Teams is a factory that initializes and returns a TeamsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Teams(accountID uuid.UUID) (api.TeamsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	return &TeamsClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: getAccountScopedURL(c.endpoint, accountID, "teams"),
	}, nil
}

// List returns a list of teams, based on the provided filter.
func (c *TeamsClient) List(ctx context.Context, names []string) ([]*api.Team, error) {
	var buf bytes.Buffer
	filterQuery := api.TeamFilter{}
	filterQuery.Teams.Name.Any = names

	if err := json.NewEncoder(&buf).Encode(&filterQuery); err != nil {
		return nil, fmt.Errorf("failed to encode filter payload data: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/filter", c.routePrefix), &buf)
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

	var teams []*api.Team
	if err := json.NewDecoder(resp.Body).Decode(&teams); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return teams, nil
}
