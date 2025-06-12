package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

	// Finally, convert the retryablehttp client to a standard http client.
	// This allows us to retain the `http.Client` interface, and avoid specifying
	// the `retryablehttp.Client` interface in our client methods.
	httpClient := retryableClient.StandardClient()

	client := &Client{hc: httpClient}

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

// obtainCsrfToken fetches the CSRF token from the Prefect server.
// It should be called after the client's endpoint and auth are configured.
func (c *Client) obtainCsrfToken() error {
	tokenURL := fmt.Sprintf("%s/csrf-token?client=%s", c.endpoint, c.csrfClientToken)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, tokenURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("error creating CSRF token request: %w", err)
	}

	// Set necessary headers. Note: Prefect-Csrf-Token is NOT sent for this request.
	setAuthorizationHeader(req, c.apiKey, c.basicAuthKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Prefect-Csrf-Client", c.csrfClientToken)

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error on CSRF token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("failed to fetch CSRF token, status: %s, body: %s", resp.Status, string(bodyBytes))
	}

	var tokenResponse api.CSRFTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return fmt.Errorf("failed to decode CSRF token response: %w", err)
	}

	if tokenResponse.Token == "" {
		return fmt.Errorf("CSRF token not found in response")
	}

	c.csrfToken = tokenResponse.Token

	return nil
}

// WithEndpoint configures the client to communicate with a self-hosted
// Prefect server or Prefect Cloud.
func WithEndpoint(endpoint string, host string) Option {
	return func(client *Client) error {
		_, err := url.Parse(endpoint)
		if err != nil {
			return fmt.Errorf("endpoint is not a valid url: %w", err)
		}

		if strings.HasSuffix(endpoint, "/") {
			return fmt.Errorf("endpoint %q must not include trailing slash", endpoint)
		}

		client.endpoint = endpoint
		client.endpointHost = host

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

// WithBasicAuthKey configures the basic auth key to use to authenticate to Prefect.
func WithBasicAuthKey(basicAuthKey string) Option {
	return func(client *Client) error {
		client.basicAuthKey = basicAuthKey

		return nil
	}
}

// WithCsrfEnabled configures the client to enable CSRF protection.
func WithCsrfEnabled(csrfEnabled bool) Option {
	return func(client *Client) error {
		if csrfEnabled {
			client.csrfClientToken = uuid.NewString()

			if err := client.obtainCsrfToken(); err != nil {
				return fmt.Errorf("failed to obtain CSRF token: %w", err)
			}
		}

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
	// If the response is empty, there was a problem with the request,
	// so try again.
	if resp == nil {
		return true, err
	}

	// If the response is a 409 (StatusConflict), that means the request
	// eventually succeeded and we don't need to make the request again.
	if resp.StatusCode == http.StatusConflict {
		return false, nil
	}

	// If the request is forbidden, no need to retry the request. Print
	// out the error and stop retrying.
	if resp.StatusCode == http.StatusForbidden {
		body, _ := io.ReadAll(resp.Body)

		return false, fmt.Errorf("status_code=%d, error=%w, body=%s", resp.StatusCode, err, body)
	}

	// Context-aware 404 handling: Skip retries for DELETE operations.
	// This prevents timing issues in acceptance tests during post-destroy plans.
	if resp.StatusCode == http.StatusNotFound {
		retry := true
		// Check if this is a DELETE operation - if so, don't retry 404s.
		if httpMethod, ok := ctx.Value(httpMethodContextKey).(string); ok && httpMethod == http.MethodDelete {
			retry = false
		}

		// For non-DELETE operations (GET, POST, PUT, PATCH), retry 404s.
		//
		// This is particularly relevant for block-related objects that are created asynchronously.
		//
		// NOTE: we encode the status code in the error object as a workaround
		// in cases where we want access to the status code on a failed client.Do() call
		// due to exhausted retries.
		//
		// go-retryablehttp does not return the response object on exhausted retries.
		//
		// https://github.com/hashicorp/go-retryablehttp/blob/main/client.go#L811-L825
		body, _ := io.ReadAll(resp.Body)

		return retry, fmt.Errorf("status_code=%d, error=%w, body=%s", resp.StatusCode, err, body)
	}

	// Fall back to the default retry policy for any other status codes.
	//nolint:wrapcheck // we've extended this method, no need to wrap error
	return retryablehttp.ErrorPropagatedRetryPolicy(ctx, resp, err)
}
