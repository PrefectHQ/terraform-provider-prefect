package api

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// AccessControlList is a custom type that represents
// an API response where the value can be:
// []uuid.UUID - a list of IDs
// ["*"] - a wildcard, meaning "all".
//
// nolint:musttag // we have custom marshal/unmarshal logic for this type
type AccessControlList struct {
	IDs []uuid.UUID
	All bool
}

// Custom JSON marshaling for AccessControlList
// so we can return []uuid.UUID or ["*"] back to the API.
func (acl AccessControlList) MarshalJSON() ([]byte, error) {
	if acl.All {
		data, err := json.Marshal([]string{"*"})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal wildcard ACL: %w", err)
		}

		return data, nil
	}

	data, err := json.Marshal(acl.IDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ID slice ACL: %w", err)
	}

	return data, nil
}

// Custom JSON unmarshaling for AccessControlList
// so we can accept []uuid.UUID or ["*"] from the API
// in a structured format.
func (acl *AccessControlList) UnmarshalJSON(data []byte) error {
	var ids []uuid.UUID
	if err := json.Unmarshal(data, &ids); err == nil {
		acl.IDs = ids
		acl.All = false

		return nil
	}

	var all []string
	if err := json.Unmarshal(data, &all); err == nil && len(all) == 1 && all[0] == "*" {
		acl.All = true
		acl.IDs = nil

		return nil
	}

	return errors.New("invalid AccessControlList format")
}

// AccessActorType represents an enum of type values
// used in our Access APIs.
type AccessActorType string

const (
	UserAccessor           AccessActorType = "user"
	ServiceAccountAccessor AccessActorType = "service_account"
	TeamAccessor           AccessActorType = "team"
	AllAccessors           AccessActorType = "*"
)

type ObjectActorAccess struct {
	ID    AccessControlList `json:"id"`
	Name  string            `json:"name"`
	Email *string           `json:"email"`
	Type  AccessActorType   `json:"type"`
}
