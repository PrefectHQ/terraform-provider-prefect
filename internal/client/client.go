package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

var _ = api.PrefectClient(&Client{})

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

	// By default, retryablehttp will only retry requests if there was some kind
	// of transient server or networking error. We can be more specific with this
	// by providing a custom function for determining whether or not to retry.
	retryableClient.CheckRetry = checkRetryPolicy

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

func checkRetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// If the response is a 409 (StatusConflict), that means the request
	// eventually succeeded and we don't need to make the request again.
	if resp.StatusCode == http.StatusConflict {
		return false, nil
	}

	// If the response is a 404 (NotFound), try again. This is particularly
	// relevant for block-related objects that are created asynchronously.
	if resp.StatusCode == http.StatusNotFound {
		return true, err
	}

	// Fall back to the default retry policy for any other status codes.
	//nolint:wrapcheck // we've extended this method, no need to wrap error
	return retryablehttp.DefaultRetryPolicy(ctx, resp, err)
}
