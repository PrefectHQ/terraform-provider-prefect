package api

import (
	"context"

	"github.com/google/uuid"
)

type AccountRolesClient interface {
	Get(ctx context.Context, roleID uuid.UUID) (*AccountRole, error)
	List(ctx context.Context, roleNames []string) ([]*AccountRole, error)
}

// AccountRole is a representation of an account role.
type AccountRole struct {
	BaseModel
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`

	AccountID    *uuid.UUID `json:"account_id"`
	IsSystemRole bool       `json:"is_system_role"`
}

// AccountRoleFilter defines the search filter payload
// when searching for workspace roles by name.
// example request payload:
// {"account_roles": {"name": {"any_": ["test"]}}}.
type AccountRoleFilter struct {
	AccountRoles struct {
		Name struct {
			Any []string `json:"any_"`
		} `json:"name"`
	} `json:"account_roles"`
}
