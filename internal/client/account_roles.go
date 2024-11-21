package client

import (
	"context"
	"fmt"
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
		routePrefix: getAccountScopedURL(c.endpoint, accountID, "account_roles"),
	}, nil
}

// List returns a list of account roles, based on the provided filter.
func (c *AccountRolesClient) List(ctx context.Context, roleNames []string) ([]*api.AccountRole, error) {
	filterQuery := api.AccountRoleFilter{}
	filterQuery.AccountRoles.Name.Any = roleNames

	cfg := requestConfig{
		url:          fmt.Sprintf("%s/filter", c.routePrefix),
		method:       http.MethodPost,
		body:         filterQuery,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOK,
	}

	var accountRoles []*api.AccountRole
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &accountRoles); err != nil {
		return nil, err
	}

	return accountRoles, nil
}

// Get returns an account role by ID.
func (c *AccountRolesClient) Get(ctx context.Context, roleID uuid.UUID) (*api.AccountRole, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          fmt.Sprintf("%s/%s", c.routePrefix, roleID.String()),
		body:         http.NoBody,
		successCodes: successCodesStatusOK,
		apiKey:       c.apiKey,
	}

	var accountRole api.AccountRole
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &accountRole); err != nil {
		return nil, err
	}

	return &accountRole, nil
}
