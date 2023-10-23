package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.AccountsClient(&AccountsClient{})

// AccountsClient is a client for working with accounts.
type AccountsClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}

// Accounts returns an AccountsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Accounts() (api.AccountsClient, error) {
	return &AccountsClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: fmt.Sprintf("%s/api/accounts", c.endpoint),
	}, nil
}

// Get returns details for an account by ID.
func (c *AccountsClient) Get(ctx context.Context, accountID uuid.UUID) (*api.AccountResponse, error) {
	path := fmt.Sprintf("%s/%s", c.routePrefix, accountID.String())
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

	var account api.AccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &account, nil
}

// Update modifies an existing account by ID.
func (c *AccountsClient) Update(ctx context.Context, accountID uuid.UUID, data api.AccountUpdate) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	path := fmt.Sprintf("%s/%s", c.routePrefix, accountID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, path, &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, c.apiKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code %s", resp.Status)
	}

	return nil
}

// Delete removes an account by ID.
func (c *AccountsClient) Delete(ctx context.Context, accountID uuid.UUID) error {
	path := fmt.Sprintf("%s/%s", c.routePrefix, accountID.String())
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, path, http.NoBody)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, c.apiKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code %s", resp.Status)
	}

	return nil
}
