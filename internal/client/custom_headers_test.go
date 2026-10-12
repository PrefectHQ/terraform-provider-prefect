package client_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
	"github.com/prefecthq/terraform-provider-prefect/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	customHeaderName  = "X-Custom-Header"
	customHeaderValue = "custom-value"
)

// newHeaderRecordingServer starts a test server that records the custom header
// value for every request and responds with a minimal valid payload.
func newHeaderRecordingServer(t *testing.T) (*httptest.Server, func() []string) {
	t.Helper()

	var (
		mu      sync.Mutex
		headers []string
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		headers = append(headers, r.Header.Get(customHeaderName))
		mu.Unlock()

		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case strings.HasSuffix(r.URL.Path, "/filter"):
			_, _ = w.Write([]byte(`[]`))
		case r.Method == http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{}`))
		default:
			_, _ = w.Write([]byte(`{}`))
		}
	}))
	t.Cleanup(server.Close)

	return server, func() []string {
		mu.Lock()
		defer mu.Unlock()

		return append([]string(nil), headers...)
	}
}

func newClientWithCustomHeaders(t *testing.T, endpoint string) *client.Client {
	t.Helper()

	c, err := client.New(
		client.WithEndpoint(endpoint, ""),
		client.WithAPIKey("test-api-key"),
		client.WithCustomHeaders(map[string]string{customHeaderName: customHeaderValue}),
	)
	require.NoError(t, err)

	return c
}

func assertAllRequestsHaveCustomHeader(t *testing.T, headers []string, expectedRequests int) {
	t.Helper()

	require.Len(t, headers, expectedRequests)
	for i, value := range headers {
		assert.Equal(t, customHeaderValue, value, "request #%d is missing the custom header", i)
	}
}

func TestBlockSchemaClient_SendsCustomHeaders(t *testing.T) {
	t.Parallel()

	server, recorded := newHeaderRecordingServer(t)
	c := newClientWithCustomHeaders(t, server.URL)

	blockSchemas, err := c.BlockSchemas(uuid.Nil, uuid.Nil)
	require.NoError(t, err)

	ctx := context.Background()
	id := uuid.New()

	_, err = blockSchemas.List(ctx, []uuid.UUID{id})
	require.NoError(t, err)
	_, err = blockSchemas.Create(ctx, &api.BlockSchemaCreate{})
	require.NoError(t, err)
	_, err = blockSchemas.Read(ctx, id)
	require.NoError(t, err)
	require.NoError(t, blockSchemas.Delete(ctx, id))

	assertAllRequestsHaveCustomHeader(t, recorded(), 4)
}

func TestServiceAccountsClient_SendsCustomHeaders(t *testing.T) {
	t.Parallel()

	server, recorded := newHeaderRecordingServer(t)
	c := newClientWithCustomHeaders(t, server.URL)

	serviceAccounts, err := c.ServiceAccounts(uuid.Nil)
	require.NoError(t, err)

	ctx := context.Background()
	id := uuid.NewString()

	_, err = serviceAccounts.Create(ctx, api.ServiceAccountCreateRequest{})
	require.NoError(t, err)
	_, err = serviceAccounts.List(ctx, nil)
	require.NoError(t, err)
	_, err = serviceAccounts.Get(ctx, id)
	require.NoError(t, err)
	require.NoError(t, serviceAccounts.Update(ctx, id, api.ServiceAccountUpdateRequest{}))
	require.NoError(t, serviceAccounts.Delete(ctx, id))
	_, err = serviceAccounts.RotateKey(ctx, id, api.ServiceAccountRotateKeyRequest{})
	require.NoError(t, err)

	assertAllRequestsHaveCustomHeader(t, recorded(), 6)
}
