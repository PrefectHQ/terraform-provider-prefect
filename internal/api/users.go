package api

import (
	"context"
)

// UsersClient is a client for working with users.
type UsersClient interface {
	Read(ctx context.Context, userID string) (*User, error)
	Update(ctx context.Context, userID string, payload UserUpdate) error
	Delete(ctx context.Context, userID string) error
}

// User is a client for working with users.
type User struct {
	BaseModel

	ActorID string `json:"actor_id"`

	Handle    string `json:"handle"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}

// UserUpdate is a payload for updating a user.
type UserUpdate struct {
	Handle    string `json:"handle"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
}
