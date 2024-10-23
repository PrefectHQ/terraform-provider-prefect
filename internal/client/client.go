package client

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

var _ = api.PrefectClient(&Client{})

// New creates and returns new client instance.
func New(opts ...Option) (*Client, error) {
	client := &Client{
		hc: retryablehttp.NewClient(),
	}

	var errs []error
	for _, opt := range opts {
		err := opt(client)
		// accumulate errors and return them all at once
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return client, nil
}

// MustNew returns a new client or panics if an error occurred.
func MustNew(opts ...Option) *Client {
	client, err := New(opts...)
	if err != nil {
		panic(fmt.Sprintf("error occurred during construction: %s", err))
	}

	return client
}

// WithClient configures the underlying http.Client used to send
// requests.
func WithClient(httpClient *retryablehttp.Client) Option {
	return func(client *Client) error {
		client.hc = httpClient

		return nil
	}
}

// WithEndpoint configures the client to communicate with a self-hosted
// Prefect server or Prefect Cloud.
func WithEndpoint(endpoint string) Option {
	return func(client *Client) error {
		_, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("endpoint is not a valid url: %w", err)
		}

		if strings.HasSuffix(endpoint, "/") {
			return fmt.Errorf("endpoint %q must not include trailing slash", endpoint)
		}

		client.endpoint = endpoint

		return nil
	}
}

// WithAPIKey configures the API Key to use to authenticate to Prefect.
func WithAPIKey(apiKey string) Option {
	return func(client *Client) error {
		client.apiKey = apiKey

		return nil
	}
}

// WithDefaults configures the default account and workspace ID.
func WithDefaults(accountID uuid.UUID, workspaceID uuid.UUID) Option {
	return func(client *Client) error {
		if accountID == uuid.Nil && workspaceID != uuid.Nil {
			return fmt.Errorf("an accountID must be set if a workspaceID is set: accountID is %q and workspaceID is %q", accountID, workspaceID)
		}

		client.defaultAccountID = accountID
		client.defaultWorkspaceID = workspaceID

		return nil
	}
}
