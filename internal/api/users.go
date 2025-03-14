package api

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UsersClient is a client for working with users.
type UsersClient interface {
	Read(ctx context.Context, userID string) (*User, error)
	Update(ctx context.Context, userID string, payload UserUpdate) error
	CreateAPIKey(ctx context.Context, userID string, payload UserAPIKeyCreate) (*UserAPIKey, error)
	ReadAPIKey(ctx context.Context, userID string, keyID string) (*UserAPIKey, error)
	DeleteAPIKey(ctx context.Context, userID string, keyID string) error
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

type UserAPIKey struct {
	ID         uuid.UUID  `json:"id"`
	Created    time.Time  `json:"created"`
	Name       string     `json:"name"`
	Expiration *time.Time `json:"expiration"`
	Key        string     `json:"key"`
}

type UserAPIKeyCreate struct {
	Name       string  `json:"name"`
	Expiration *string `json:"expiration"`
}
