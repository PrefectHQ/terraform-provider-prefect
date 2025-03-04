package api

import (
	"context"

	"github.com/google/uuid"
)

// TeamAccessClient is a client for the TeamAccess resource.
type TeamAccessClient interface {
	Read(ctx context.Context, teamID, memberID, memberActorID uuid.UUID) (*TeamAccess, error)
	Upsert(ctx context.Context, memberType string, memberID uuid.UUID) error
	Delete(ctx context.Context, memberID uuid.UUID) error
}

// TeamAccess is a representation of a team access.
type TeamAccess struct {
	BaseModel

	TeamID uuid.UUID `json:"team_id"`

	MemberID      uuid.UUID `json:"member_id"`
	MemberActorID uuid.UUID `json:"member_actor_id"`
	MemberType    string    `json:"member_type"`
}

// TeamAccessUpsert defines the payload for an upsert request.
type TeamAccessUpsert struct {
	Members []TeamAccessMember `json:"members"`
}

// TeamAccessMember is a representation of a team access member.
type TeamAccessMember struct {
	MemberID   uuid.UUID `json:"member_id"`
	MemberType string    `json:"member_type"`
}

// TeamAccessRead defines the response payload for a get request.
type TeamAccessRead struct {
	Memberships []Membership `json:"memberships"`
}

// Membership is a representation of a team access membership.
type Membership struct {
	ActorID uuid.UUID `json:"id"`
	Type    string    `json:"type"`
}
