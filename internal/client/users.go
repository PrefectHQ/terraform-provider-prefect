package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.UsersClient(&UsersClient{})

type UsersClient struct {
	hc           *http.Client
	apiKey       string
	basicAuthKey string
	routePrefix  string
}

// Users is a factory that initializes and returns a UsersClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Users() (api.UsersClient, error) {
	return &UsersClient{
		hc:           c.hc,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		routePrefix:  fmt.Sprintf("%s/users", c.endpoint),
	}, nil
}

// Read reads a user.
func (c *UsersClient) Read(ctx context.Context, userID string) (*api.User, error) {
	cfg := requestConfig{
		url:          fmt.Sprintf("%s/%s", c.routePrefix, userID),
		method:       http.MethodGet,
		body:         http.NoBody,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusOK,
	}

	var user api.User
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Update updates a user.
func (c *UsersClient) Update(ctx context.Context, userID string, payload api.UserUpdate) (*api.User, error) {
	cfg := requestConfig{
		url:          fmt.Sprintf("%s/%s", c.routePrefix, userID),
		method:       http.MethodPatch,
		body:         payload,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return nil, nil
}

// Delete deletes a user.
func (c *UsersClient) Delete(ctx context.Context, userID string) error {
	cfg := requestConfig{
		url:          fmt.Sprintf("%s/%s", c.routePrefix, userID),
		method:       http.MethodDelete,
		body:         http.NoBody,
		apiKey:       c.apiKey,
		basicAuthKey: c.basicAuthKey,
		successCodes: successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
