package client

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

var _ = api.PrefectClient(&Client{})

const (
	//nolint:revive // matches name from retryablehttp package
	clientRetryWaitMin time.Duration = 3 * time.Second
	clientRetryMax     int           = 10
)

// New creates and returns new client instance.
func New(opts ...Option) (*Client, error) {
	// Uses the retryablehttp package for built-in retries
	// with exponential backoff.
	//
	// Some notable defaults from that package include:
	// - max retries: 4
	// - retry wait minimum seconds: 1
	// - retry wait maximum seconds: 30
	//
	// All defaults are defined in
	// https://github.com/hashicorp/go-retryablehttp/blob/main/client.go#L48-L51.
	retryableClient := retryablehttp.NewClient()

	// We adjust the seconds between retries and the maximum number of retries.
	// This provides a bigger window of time for the API to return the desired
	// response retries. This is especially relevant for Block-related objects
	// that are created asynchronously after a new Workspace request resolves.
	retryableClient.RetryWaitMin = clientRetryWaitMin
	retryableClient.RetryMax = clientRetryMax

	client := &Client{
		hc: retryableClient,
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
