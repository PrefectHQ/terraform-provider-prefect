package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
func (c *Client) Accounts(accountID uuid.UUID) (api.AccountsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if accountID == uuid.Nil {
		return nil, fmt.Errorf("accountID must be set: accountID is %q", accountID)
	}

	return &AccountsClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: getAccountScopedURL(c.endpoint, accountID, ""),
	}, nil
}

// Get returns details for an account by ID.
func (c *AccountsClient) Get(ctx context.Context) (*api.AccountResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.routePrefix, http.NoBody)
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

	var account api.AccountResponse
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &account, nil
}

// Update modifies an existing account by ID.
func (c *AccountsClient) Update(ctx context.Context, data api.AccountUpdate) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.routePrefix, &buf)
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
		errorBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	return nil
}

// UpdateSettings modifies an existing account's settings by ID.
func (c *AccountsClient) UpdateSettings(ctx context.Context, data api.AccountSettingsUpdate) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data.Settings); err != nil {
		return fmt.Errorf("failed to encode data: %w", err)
	}

	url := c.routePrefix + "settings"
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, &buf)
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
		errorBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	return nil
}

// Delete removes an account by ID.
func (c *AccountsClient) Delete(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.routePrefix, http.NoBody)
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
		errorBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	return nil
}
