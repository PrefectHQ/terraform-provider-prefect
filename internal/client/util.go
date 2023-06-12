package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// getAccountScopedURL constructs a URL for an account-scoped route.
func getAccountScopedURL(endpoint string, accountID uuid.UUID, route string) string {
	var builder strings.Builder

	builder.WriteString(endpoint)

	builder.WriteString("/accounts/")
	builder.WriteString(accountID.String())
	builder.WriteRune('/')

	builder.WriteString(route)

	return builder.String()
}

// getWorkspaceScopedURL constructs a URL for a workspace-scoped route.
func getWorkspaceScopedURL(endpoint string, accountID uuid.UUID, workspaceID uuid.UUID, route string) string {
	var builder strings.Builder

	builder.WriteString(endpoint)

	if accountID != uuid.Nil && workspaceID != uuid.Nil {
		builder.WriteString("/accounts/")
		builder.WriteString(accountID.String())

		builder.WriteString("/workspaces/")
		builder.WriteString(workspaceID.String())
	}

	builder.WriteRune('/')
	builder.WriteString(route)

	return builder.String()
}

// setAuthorizationHeader will set the Authorization header to the
// provided apiKey, if set.
func setAuthorizationHeader(request *http.Request, apiKey string) {
	if apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	}
}

func doRequest(client *http.Client, apiKey string, request *http.Request) (*http.Response, error) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	setAuthorizationHeader(request, apiKey)

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	return resp, nil
}
