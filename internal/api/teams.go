package api

import (
	"context"
)

// TeamsClient is a client for working with teams.
type TeamsClient interface {
	List(ctx context.Context, names []string) ([]*Team, error)
}

// Team is a representation of an team.
type Team struct {
	BaseModel
	Name        string `json:"name"`
	Description string `json:"description"`
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
