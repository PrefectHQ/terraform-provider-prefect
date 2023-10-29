package api

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AccountMembershipsClient interface {
	List(ctx context.Context, filterQuery AccountMembershipFilter) ([]*AccountMembership, error)
}

type AccountMembership struct {
	ID              uuid.UUID  `json:"id"`
	ActorID         uuid.UUID  `json:"actor_id"`
	UserID          uuid.UUID  `json:"user_id"`
	FirstName       string     `json:"first_name"`
	LastName        string     `json:"last_name"`
	Handle          string     `json:"handle"`
	Email           string     `json:"email"`
	AccountRoleName string     `json:"account_role_name"`
	AccountRoleID   uuid.UUID  `json:"account_role_id"`
	LastLogin       *time.Time `json:"last_login"`
}

// AccountMembershipFilter defines the search filter payload
// when searching for workspace roles by name.
// example request payload:
// {"account_memberships": {"email": {"any_": ["test"]}}}.
type AccountMembershipFilter struct {
	AccountMemberships struct {
		Email *struct {
			Any []string `json:"any_"`
		} `json:"email,omitempty"`
	} `json:"account_memberships"`
}
