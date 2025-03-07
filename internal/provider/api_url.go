package provider

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

const (
	apiURLWithIDs = `^https://[^/]+/api/accounts/([a-zA-Z0-9-]+)/workspaces/([a-zA-Z0-9-]+)$`

	accountIDPathIndex  = 3
	workspacesPathIndex = 5
	expectedPathLength  = 6
)

func URLContainsIDs(url string) bool {
	return regexp.MustCompile(apiURLWithIDs).MatchString(url)
}

func GetUUIDFromPath(path string, pathIndex int) (uuid.UUID, error) {
	parts := strings.Split(path, "/")
	if len(parts) != expectedPathLength {
		return uuid.Nil, fmt.Errorf("URL path does not contain expected number of parts: %d", expectedPathLength)
	}

	u, err := uuid.Parse(parts[pathIndex])
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to parse UUID from path: %w", err)
	}

	return u, nil
}

func GetAccountIDFromPath(path string) (uuid.UUID, error) {
	return GetUUIDFromPath(path, accountIDPathIndex)
}

func GetWorkspaceIDFromPath(path string) (uuid.UUID, error) {
	return GetUUIDFromPath(path, workspacesPathIndex)
}
