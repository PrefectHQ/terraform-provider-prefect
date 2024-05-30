package api

import (
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
)

// AccessActorID is a custom type that represents
// an API response where the value can be:
// uuid.UUID - a single ID
// "*" - a wildcard string, meaning "all".
//
// nolint:musttag // we have custom marshal/unmarshal logic for this type
type AccessActorID struct {
	ID  *uuid.UUID
	All bool
}

// Custom JSON marshaling for AccessActorID
// so we can return uuid.UUID or "*" back to the API.
func (aid AccessActorID) MarshalJSON() ([]byte, error) {
	if aid.All {
		return []byte("*"), nil
	}
	if aid.ID != nil {
		uuidByteSlice, err := json.Marshal(aid.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal AccessActorID: %w", err)
		}

		return uuidByteSlice, nil
	}

	return nil, fmt.Errorf("invalid AccessActorID: both ID and All are nil/false")
}

// Custom JSON unmarshaling for AccessActorID
// so we can accept uuid.UUID or "*" from the API
// in a structured format.
func (aid *AccessActorID) UnmarshalJSON(data []byte) error {
	var id uuid.UUID
	if err := json.Unmarshal(data, &id); err == nil {
		aid.ID = &id
		aid.All = false

		return nil
	}

	var all string
	if err := json.Unmarshal(data, &all); err == nil && all == "*" {
		aid.All = true
		aid.ID = nil

		return nil
	}

	return fmt.Errorf("invalid AccessActorID format")
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
	ID    AccessActorID   `json:"id"`
	Name  string          `json:"name"`
	Email *string         `json:"email"`
	Type  AccessActorType `json:"type"`
}
