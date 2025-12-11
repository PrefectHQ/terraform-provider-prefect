package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.AccountMembershipsClient(&AccountMembershipsClient{})

type AccountMembershipsClient struct {
	hc              *http.Client
	apiKey          string
	basicAuthKey    string
	routePrefix     string
	csrfClientToken string
	csrfToken       string
	customHeaders   map[string]string
}

// AccountMemberships is a factory that initializes and returns a AccountMembershipsClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) AccountMemberships(accountID uuid.UUID) (api.AccountMembershipsClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	return &AccountMembershipsClient{
		hc:              c.hc,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		routePrefix:     getAccountScopedURL(c.endpoint, accountID, "account_memberships"),
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
	}, nil
}

// List returns a list of account memberships, based on the provided filter.
func (c *AccountMembershipsClient) List(ctx context.Context, emails []string) ([]*api.AccountMembership, error) {
	filterQuery := api.AccountMembershipFilter{}
	filterQuery.AccountMemberships.Email.Any = emails

	cfg := requestConfig{
		url:             fmt.Sprintf("%s/filter", c.routePrefix),
		method:          http.MethodPost,
		body:            filterQuery,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusOK,
	}

	var accountMemberships []*api.AccountMembership
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &accountMemberships); err != nil {
		return nil, err
	}

	return accountMemberships, nil
}

// Update updates the account membership for the given account membership ID and account role ID.
func (c *AccountMembershipsClient) Update(ctx context.Context, accountMembershipID uuid.UUID, payload *api.AccountMembershipUpdate) error {
	cfg := requestConfig{
		url:             fmt.Sprintf("%s/%s", c.routePrefix, accountMembershipID),
		method:          http.MethodPatch,
		body:            payload,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Delete deletes the account membership for the given account membership ID.
func (c *AccountMembershipsClient) Delete(ctx context.Context, accountMembershipID uuid.UUID) error {
	cfg := requestConfig{
		url:             fmt.Sprintf("%s/%s", c.routePrefix, accountMembershipID),
		method:          http.MethodDelete,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
