package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.UsersClient(&UsersClient{})

// UsersClient is a client for Prefect Users.
type UsersClient struct {
	hc              *http.Client
	apiKey          string
	basicAuthKey    string
	routePrefix     string
	csrfClientToken string
	csrfToken       string
	customHeaders   map[string]string
}

// Users is a factory that initializes and returns a UsersClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Users() (api.UsersClient, error) {
	return &UsersClient{
		hc:              c.hc,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		routePrefix:     fmt.Sprintf("%s/users", c.endpoint),
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
	}, nil
}

// Read reads a user.
func (c *UsersClient) Read(ctx context.Context, userID string) (*api.User, error) {
	cfg := requestConfig{
		url:             fmt.Sprintf("%s/%s", c.routePrefix, userID),
		method:          http.MethodGet,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusOK,
	}

	var user api.User
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

// Update updates a user.
func (c *UsersClient) Update(ctx context.Context, userID string, payload api.UserUpdate) error {
	cfg := requestConfig{
		url:             fmt.Sprintf("%s/%s", c.routePrefix, userID),
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

// CreateAPIKey creates an API key for a user.
func (c *UsersClient) CreateAPIKey(ctx context.Context, userID string, payload api.UserAPIKeyCreate) (*api.UserAPIKey, error) {
	cfg := requestConfig{
		url:             fmt.Sprintf("%s/%s/api_keys", c.routePrefix, userID),
		method:          http.MethodPost,
		body:            payload,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusCreated,
	}

	var apiKey api.UserAPIKey
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &apiKey); err != nil {
		return nil, err
	}

	return &apiKey, nil
}

// ReadAPIKey reads an API key for a user.
func (c *UsersClient) ReadAPIKey(ctx context.Context, userID string, apiKeyID string) (*api.UserAPIKey, error) {
	cfg := requestConfig{
		url:             fmt.Sprintf("%s/%s/api_keys/%s", c.routePrefix, userID, apiKeyID),
		method:          http.MethodGet,
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		customHeaders:   c.customHeaders,
		successCodes:    successCodesStatusOK,
	}

	var apiKey api.UserAPIKey
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &apiKey); err != nil {
		return nil, err
	}

	return &apiKey, nil
}

// DeleteAPIKey deletes an API key for a user.
func (c *UsersClient) DeleteAPIKey(ctx context.Context, userID string, apiKeyID string) error {
	cfg := requestConfig{
		url:             fmt.Sprintf("%s/%s/api_keys/%s", c.routePrefix, userID, apiKeyID),
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
