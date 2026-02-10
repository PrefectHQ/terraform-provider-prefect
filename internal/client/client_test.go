package client_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/prefecthq/terraform-provider-prefect/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckRetryPolicy_NilResponse(t *testing.T) {
	t.Parallel()

	retry, err := client.CheckRetryPolicy(context.Background(), nil, nil)

	assert.True(t, retry, "should retry when response is nil")
	assert.NoError(t, err)
}

func TestCheckRetryPolicy_Conflict(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusConflict,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	retry, err := client.CheckRetryPolicy(context.Background(), resp, nil)

	assert.False(t, retry, "should not retry on 409 Conflict")
	assert.NoError(t, err)
}

func TestCheckRetryPolicy_Forbidden(t *testing.T) {
	t.Parallel()

	body := `{"error": "access denied"}`
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	retry, err := client.CheckRetryPolicy(context.Background(), resp, nil)

	assert.False(t, retry, "should not retry on 403 Forbidden")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status_code=403")
	assert.Contains(t, err.Error(), "access denied")
}

func TestCheckRetryPolicy_NotFound_DELETE(t *testing.T) {
	t.Parallel()

	body := `{"detail": "not found"}`
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	// Add HTTP method to context for DELETE
	ctx := context.WithValue(context.Background(), client.HTTPMethodContextKey, http.MethodDelete)

	retry, err := client.CheckRetryPolicy(ctx, resp, nil)

	assert.False(t, retry, "should not retry 404 on DELETE")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status_code=404")
}

func TestCheckRetryPolicy_NotFound_GET(t *testing.T) {
	t.Parallel()

	body := `{"detail": "not found"}`
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	// Add HTTP method to context for GET
	ctx := context.WithValue(context.Background(), client.HTTPMethodContextKey, http.MethodGet)

	retry, err := client.CheckRetryPolicy(ctx, resp, nil)

	assert.True(t, retry, "should retry 404 on GET")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status_code=404")
}

func TestCheckRetryPolicy_NotFound_POST(t *testing.T) {
	t.Parallel()

	body := `{"detail": "not found"}`
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	// Add HTTP method to context for POST
	ctx := context.WithValue(context.Background(), client.HTTPMethodContextKey, http.MethodPost)

	retry, err := client.CheckRetryPolicy(ctx, resp, nil)

	assert.True(t, retry, "should retry 404 on POST")
	assert.Error(t, err)
}

func TestCheckRetryPolicy_TooManyRequests(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	retry, err := client.CheckRetryPolicy(context.Background(), resp, nil)

	// Falls back to ErrorPropagatedRetryPolicy which should retry 429
	assert.True(t, retry, "should retry on 429 Too Many Requests")
	assert.NoError(t, err)
}

func TestCheckRetryPolicy_ServiceUnavailable(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	retry, err := client.CheckRetryPolicy(context.Background(), resp, nil)

	// Falls back to ErrorPropagatedRetryPolicy which should retry 503
	// ErrorPropagatedRetryPolicy propagates the error (unlike DefaultRetryPolicy)
	assert.True(t, retry, "should retry on 503 Service Unavailable")
	// Error is expected - ErrorPropagatedRetryPolicy returns an error for 5xx
	assert.Error(t, err)
}

func TestCheckRetryPolicy_Success(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}

	retry, err := client.CheckRetryPolicy(context.Background(), resp, nil)

	assert.False(t, retry, "should not retry on 200 OK")
	assert.NoError(t, err)
}

// TestClientCreation_Success verifies basic client creation works.
func TestClientCreation_Success(t *testing.T) {
	t.Parallel()

	c, err := client.New()

	require.NoError(t, err)
	assert.NotNil(t, c)
	assert.NotNil(t, c.HTTPClient())
}

// TestClientCreation_WithOptions verifies client creation with options.
func TestClientCreation_WithOptions(t *testing.T) {
	t.Parallel()

	c, err := client.New(
		client.WithEndpoint("https://api.prefect.cloud", "api.prefect.cloud"),
		client.WithAPIKey("test-api-key"),
	)

	require.NoError(t, err)
	assert.NotNil(t, c)
	assert.Equal(t, "https://api.prefect.cloud", c.Endpoint())
	assert.Equal(t, "test-api-key", c.APIKey())
}

// TestClientCreation_InvalidEndpoint verifies error handling for invalid endpoints.
func TestClientCreation_InvalidEndpoint(t *testing.T) {
	t.Parallel()

	c, err := client.New(
		client.WithEndpoint("https://api.prefect.cloud/", "api.prefect.cloud"),
	)

	require.Error(t, err)
	assert.Nil(t, c)
	assert.Contains(t, err.Error(), "must not include trailing slash")
}

// TestCheckRetryPolicy_NotFound_NoMethod tests 404 handling when no HTTP method is in context.
func TestCheckRetryPolicy_NotFound_NoMethod(t *testing.T) {
	t.Parallel()

	body := `{"detail": "not found"}`
	resp := &http.Response{
		StatusCode: http.StatusNotFound,
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	// No HTTP method in context
	retry, err := client.CheckRetryPolicy(context.Background(), resp, nil)

	// Should retry since it's not a DELETE
	assert.True(t, retry, "should retry 404 when method is not DELETE")
	assert.Error(t, err)
}

// TestCheckRetryPolicy_WithBodyReset tests that the response body can be read after policy check.
// This is important because client.CheckRetryPolicy reads the body for error messages.
func TestCheckRetryPolicy_BodyReadBehavior(t *testing.T) {
	t.Parallel()

	body := `{"detail": "forbidden access"}`
	resp := &http.Response{
		StatusCode: http.StatusForbidden,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
	}

	_, err := client.CheckRetryPolicy(context.Background(), resp, nil)

	// The error should contain the body content
	require.Error(t, err)
	assert.Contains(t, err.Error(), "forbidden access")
}
