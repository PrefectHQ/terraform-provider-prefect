package api

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
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Email *string         `json:"email"`
	Type  AccessActorType `json:"type"`
}
