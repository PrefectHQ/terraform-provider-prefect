package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.BotsClient(&BotsClient{})

// BotsClient is a client for working with service accounts.
type BotsClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}

// Bots returns a BotsClient.
func (c *Client) Bots() (api.BotsClient, error) {
	return &BotsClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: fmt.Sprintf("%s/api/accounts/%s/bots", c.endpoint, c.defaultAccountID),
	}, nil
}

// Get a single bot by ID
func (c *BotsClient) Get(ctx context.Context, id uuid.UUID) (*api.Bot, error) {
	path := fmt.Sprintf("%s/%s", c.routePrefix, id.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, http.NoBody)
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
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var bot api.Bot
	if err := json.NewDecoder(resp.Body).Decode(&bot); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &bot, nil
}
