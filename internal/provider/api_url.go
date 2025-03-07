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
)

func urlContainsIDs(url string) bool {
	return regexp.MustCompile(apiURLWithIDs).MatchString(url)
}

func getUUIDFromPath(path string, pathIndex int) (uuid.UUID, error) {
	parts := strings.Split(path, "/")
	u, err := uuid.Parse(parts[pathIndex])
	if err != nil {
		return uuid.Nil, fmt.Errorf("unable to parse workspace ID from path: %w", err)
	}

	return u, nil
}

func getAccountIDFromPath(path string) (uuid.UUID, error) {
	return getUUIDFromPath(path, accountIDPathIndex)
}

func getWorkspaceIDFromPath(path string) (uuid.UUID, error) {
	return getUUIDFromPath(path, workspacesPathIndex)
}
