package api

import (
	"context"
)

// TeamsClient is a client for working with teams.
type TeamsClient interface {
	Create(ctx context.Context, payload TeamCreate) (*Team, error)
	Read(ctx context.Context, teamID string) (*Team, error)
	List(ctx context.Context, names []string) ([]*Team, error)
	Update(ctx context.Context, teamID string, payload TeamUpdate) (*Team, error)
	Delete(ctx context.Context, teamID string) error
}

// Team is a representation of an team.
type Team struct {
	BaseModel
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// TeamCreate is a payload for creating a team.
type TeamCreate struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// TeamUpdate is a payload for updating a team.
type TeamUpdate struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// TeamFilter defines the search filter payload
// when searching for team by name.
// example request payload:
// {"teams": {"name": {"any_": ["test"]}}}.
type TeamFilter struct {
	Teams struct {
		Name struct {
			Any []string `json:"any_"`
		} `json:"name"`
	} `json:"teams"`
}

// TeamFilterRequest wraps TeamFilter with pagination parameters
// for the POST /teams/filter endpoint.
type TeamFilterRequest struct {
	TeamFilter
	Limit  *int64 `json:"limit,omitempty"`
	Offset *int64 `json:"offset,omitempty"`
}
