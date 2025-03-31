package api

import (
	"context"

	"github.com/google/uuid"
)

// FlowsClient is a client for working with flows.
type FlowsClient interface {
	Create(ctx context.Context, data FlowCreate) (*Flow, error)
	Get(ctx context.Context, flowID uuid.UUID) (*Flow, error)
	List(ctx context.Context, handleNames []string) ([]*Flow, error)
	Update(ctx context.Context, flowID uuid.UUID, data FlowUpdate) error
	Delete(ctx context.Context, flowID uuid.UUID) error
}

// Flow is a representation of a flow.
type Flow struct {
	BaseModel
	AccountID   uuid.UUID `json:"account_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
	Name        string    `json:"name"`
	Tags        []string  `json:"tags"`
}

// FlowCreate is a subset of Flow used when creating flows.
type FlowCreate struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

// FlowUpdate is a subset of Flow used when updating flows.
type FlowUpdate struct {
	Tags []string `json:"tags"`
}

// FlowFilter defines the search filter payload
// when searching for flows by name.
// example request payload:
// {"flows": {"handle": {"any_": ["test"]}}}.
type FlowFilter struct {
	Flows struct {
		Handle struct {
			Any []string `json:"any_"`
		} `json:"handle"`
	} `json:"flows"`
}
