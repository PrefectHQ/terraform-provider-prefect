package api

import (
	"context"

	"github.com/google/uuid"
)

// WorkPoolsClient is a client for working with work pools.
type WorkPoolsClient interface {
	Get(ctx context.Context, name string) (*WorkPool, error)
	Create(ctx context.Context, data WorkPoolCreate) (*WorkPool, error)
	Update(ctx context.Context, data WorkPoolUpdate) error
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
	Name             string                 `json:"name"`
	Description      *string                `json:"description"`
	Type             string                 `json:"type"`
	BaseJobTemplate  map[string]interface{} `json:"base_job_template"`
	IsPaused         bool                   `json:"is_paused"`
	ConcurrencyLimit *int64                 `json:"concurrency_limit"`
}

// WorkPoolUpdate is a subset of WorkPool used when updating pools.
type WorkPoolUpdate struct {
	Description      *string                `json:"description"`
	IsPaused         *bool                  `json:"is_paused"`
	BaseJobTemplate  map[string]interface{} `json:"base_job_template"`
	ConcurrencyLimit *int64                 `json:"concurrency_limit"`
}
