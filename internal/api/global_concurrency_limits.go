package api

import (
	"context"

	"github.com/google/uuid"
)

// GlobalConcurrencyLimitsClient is a client for working with global concurrency limits.
type GlobalConcurrencyLimitsClient interface {
	Create(ctx context.Context, globalConcurrencyLimit GlobalConcurrencyLimitCreate) (*GlobalConcurrencyLimit, error)
	Read(ctx context.Context, globalConcurrencyLimitID string) (*GlobalConcurrencyLimit, error)
	Update(ctx context.Context, globalConcurrencyLimitID string, globalConcurrencyLimit GlobalConcurrencyLimitUpdate) (*GlobalConcurrencyLimit, error)
	Delete(ctx context.Context, globalConcurrencyLimitID string) error
}

// GlobalConcurrencyLimit is a representation of a global concurrency limit.
type GlobalConcurrencyLimit struct {
	BaseModel
	Active             bool   `json:"active"`
	Name               string `json:"name"`
	Limit              int64  `json:"limit"`
	ActiveSlots        int64  `json:"active_slots"`
	DeniedSlots        int64  `json:"denied_slots"`
	SlotDecayPerSecond int64  `json:"slot_decay_per_second"`
}

// GlobalConcurrencyLimitCreate is a subset of GlobalConcurrencyLimit used when creating global concurrency limits.
type GlobalConcurrencyLimitCreate struct {
	Active             bool   `json:"active"`
	Name               string `json:"name"`
	Limit              int64  `json:"limit"`
	ActiveSlots        int64  `json:"active_slots"`
	DeniedSlots        int64  `json:"denied_slots"`
	SlotDecayPerSecond int64  `json:"slot_decay_per_second"`
}

// GlobalConcurrencyLimitUpdate is a subset of GlobalConcurrencyLimit used when updating global concurrency limits.
type GlobalConcurrencyLimitUpdate struct {
	Active             bool   `json:"active"`
	Name               string `json:"name"`
	Limit              int64  `json:"limit"`
	ActiveSlots        int64  `json:"active_slots"`
	DeniedSlots        int64  `json:"denied_slots"`
	SlotDecayPerSecond int64  `json:"slot_decay_per_second"`
}

// GlobalConcurrencyLimitFilter is a filter for global concurrency limits.
type GlobalConcurrencyLimitFilter struct {
	Any []uuid.UUID `json:"any_"`
}
