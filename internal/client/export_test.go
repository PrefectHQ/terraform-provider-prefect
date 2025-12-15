package client

import "net/http"

// Export internal functions and types for testing.
// This file is only compiled during tests.

// CheckRetryPolicy exports checkRetryPolicy for testing.
var CheckRetryPolicy = checkRetryPolicy

// HTTPMethodContextKey exports httpMethodContextKey for testing.
const HTTPMethodContextKey = httpMethodContextKey

// HTTPClient returns the internal HTTP client for testing purposes.
func (c *Client) HTTPClient() *http.Client {
	return c.hc
}

// Endpoint returns the endpoint for testing purposes.
func (c *Client) Endpoint() string {
	return c.endpoint
}

// APIKey returns the API key for testing purposes.
func (c *Client) APIKey() string {
	return c.apiKey
}
