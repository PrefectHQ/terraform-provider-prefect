package api

import (
	"context"

	"github.com/google/uuid"
)

// WorkPoolsClient is a client for working with work pools.
type WorkPoolsClient interface {
	Create(ctx context.Context, data WorkPoolCreate) (*WorkPool, error)
	List(ctx context.Context, filter WorkPoolFilter) ([]*WorkPool, error)
	Get(ctx context.Context, name string) (*WorkPool, error)
	Update(ctx context.Context, name string, data WorkPoolUpdate) error
	Delete(ctx context.Context, name string) error
}

// WorkPool is a representation of a work pool.
type WorkPool struct {
	BaseModel
	Name             string                 `json:"name"`
	Description      *string                `json:"description"`
	Type             string                 `json:"type"`
	BaseJobTemplate  map[string]interface{} `json:"base_job_template"`
	IsPaused         bool                   `json:"is_paused"`
	ConcurrencyLimit *int64                 `json:"concurrency_limit"`
	DefaultQueueID   uuid.UUID              `json:"default_queue_id"`
}

// WorkPoolCreate is a subset of WorkPool used when creating pools.
type WorkPoolCreate struct {
	Name             string                  `json:"name"`
	Description      *string                 `json:"description"`
	Type             string                  `json:"type"`
	BaseJobTemplate  *map[string]interface{} `json:"base_job_template,omitempty"`
	IsPaused         bool                    `json:"is_paused"`
	ConcurrencyLimit *int64                  `json:"concurrency_limit"`
}

// WorkPoolUpdate is a subset of WorkPool used when updating pools.
type WorkPoolUpdate struct {
	Description      *string                 `json:"description"`
	IsPaused         *bool                   `json:"is_paused"`
	BaseJobTemplate  *map[string]interface{} `json:"base_job_template,omitempty"`
	ConcurrencyLimit *int64                  `json:"concurrency_limit"`
}

// WorkPoolFilter defines filters when searching for work pools.
type WorkPoolFilter struct {
	Any []uuid.UUID `json:"any_"`
}
