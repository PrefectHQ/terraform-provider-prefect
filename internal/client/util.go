package client

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// getAccountScopedURL constructs a URL for an account-scoped route
func getAccountScopedURL(endpoint string, accountID uuid.UUID, route string) string {
	var sb strings.Builder
	sb.WriteString(endpoint)
	sb.WriteString("/accounts/")
	sb.WriteString(accountID.String())
	sb.WriteRune('/')
	sb.WriteString(route)
	return sb.String()
}

// getWorkspaceScopedURL constructs a URL for a workspace-scoped route
func getWorkspaceScopedURL(endpoint string, accountID uuid.UUID, workspaceID uuid.UUID, route string) string {
	var sb strings.Builder
	sb.WriteString(endpoint)
	if accountID != uuid.Nil && workspaceID != uuid.Nil {
		sb.WriteString("/accounts/")
		sb.WriteString(accountID.String())
		sb.WriteString("/workspaces/")
		sb.WriteString(workspaceID.String())
	}
	sb.WriteRune('/')
	sb.WriteString(route)
	return sb.String()
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

	if apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	return resp, nil
}
