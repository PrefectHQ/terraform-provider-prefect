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

var _ = api.AccountMembershipsClient(&AccountMembershipsClient{})

type AccountMembershipsClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}

// AccountMemberships is a factory that initializes and returns a AccountMembershipsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) AccountMemberships(accountID uuid.UUID) (api.AccountMembershipsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	return &AccountMembershipsClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: fmt.Sprintf("%s/accounts/%s/account_memberships", c.endpoint, accountID.String()),
	}, nil
}

// List returns a list of account memberships, based on the provided filter.
func (c *AccountMembershipsClient) List(ctx context.Context, filterQuery api.AccountMembershipFilter) ([]*api.AccountMembership, error) {
	var buf bytes.Buffer

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

	var accountMemberships []*api.AccountMembership
	if err := json.NewDecoder(resp.Body).Decode(&accountMemberships); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return accountMemberships, nil
}
