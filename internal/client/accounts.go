package client

import (
	"context"
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
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          c.routePrefix,
		body:         http.NoBody,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOK,
	}

	var account api.AccountResponse
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &account); err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}

	return &account, nil
}

// Update modifies an existing account by ID.
func (c *AccountsClient) Update(ctx context.Context, data api.AccountUpdate) error {
	cfg := requestConfig{
		method:       http.MethodPatch,
		url:          c.routePrefix,
		body:         data,
		apiKey:       c.apiKey,
		successCodes: []int{http.StatusOK, http.StatusNoContent},
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// UpdateSettings modifies an existing account's settings by ID.
func (c *AccountsClient) UpdateSettings(ctx context.Context, data api.AccountSettingsUpdate) error {
	cfg := requestConfig{
		method:       http.MethodPatch,
		url:          c.routePrefix + "settings",
		body:         data.AccountSettings,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update account settings: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Delete removes an account by ID.
func (c *AccountsClient) Delete(ctx context.Context) error {
	cfg := requestConfig{
		method:       http.MethodDelete,
		url:          c.routePrefix,
		body:         http.NoBody,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
