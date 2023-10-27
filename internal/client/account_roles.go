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

// type assertion ensures that this client implements the interface.
var _ = api.AccountRolesClient(&AccountRolesClient{})

type AccountRolesClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}

// AccountRoles is a factory that initializes and returns a AccountRolesClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) AccountRoles(accountID uuid.UUID) (api.AccountRolesClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	return &AccountRolesClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: fmt.Sprintf("%s/accounts/%s/account_roles", c.endpoint, accountID.String()),
	}, nil
}

// list returns a list of account roles, based on the provided filter.
func (c *AccountRolesClient) List(ctx context.Context, roleNames []string) ([]*api.AccountRole, error) {
	var buf bytes.Buffer
	filterQuery := api.AccountRoleFilter{}
	filterQuery.AccountRoles.Name.Any = roleNames

	if err := json.NewEncoder(&buf).Encode(&filterQuery); err != nil {
		return nil, fmt.Errorf("failed to encode filter payload data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/filter", c.routePrefix), &buf)
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

	var accountRoles []*api.AccountRole
	if err := json.NewDecoder(resp.Body).Decode(&accountRoles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return accountRoles, nil
}

// Get returns an account role by ID.
func (c *AccountRolesClient) Get(ctx context.Context, roleID uuid.UUID) (*api.AccountRole, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s", c.routePrefix, roleID.String()), http.NoBody)
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

	var accountRole api.AccountRole
	if err := json.NewDecoder(resp.Body).Decode(&accountRole); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &accountRole, nil
}
