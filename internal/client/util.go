package client

import (
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/go-retryablehttp"
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
func setAuthorizationHeader(request *retryablehttp.Request, apiKey string) {
	if apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	}
}

// setDefaultHeaders will set Authorization, Content-Type, and Accept
// headers that are common to most requests.
func setDefaultHeaders(request *retryablehttp.Request, apiKey string) {
	setAuthorizationHeader(request, apiKey)

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
}
