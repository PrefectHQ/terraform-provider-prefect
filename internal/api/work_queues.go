package api

import (
	"context"

	"github.com/google/uuid"
)

// WorkQueuesClient is a client for working with work queues.
type WorkQueuesClient interface {
	Create(ctx context.Context, data WorkQueueCreate) (*WorkQueue, error)
	List(ctx context.Context, filter WorkQueueFilter) ([]*WorkQueue, error)
	Get(ctx context.Context, name string) (*WorkQueue, error)
	Update(ctx context.Context, name string, data WorkQueueUpdate) error
	Delete(ctx context.Context, name string) error
}

// WorkQueue is a representation of a work queue.
type WorkQueue struct {
	BaseModel
	Name             string    `json:"name"`
	WorkPoolName     string    `json:"work_pool_name"`
	Description      *string   `json:"description"`
	IsPaused         bool      `json:"is_paused"`
	ConcurrencyLimit *int64    `json:"concurrency_limit"`
	Priority         *int64    `json:"priority"`
	QueueID          uuid.UUID `json:"queue_id"`
}

// WorkQueueCreate is a subset of WorkQueue used when creating queues.
type WorkQueueCreate struct {
	Name             string  `json:"name"`
	Description      *string `json:"description"`
	IsPaused         bool    `json:"is_paused"`
	ConcurrencyLimit *int64  `json:"concurrency_limit"`
	Priority         *int64  `json:"priority"`
}

// WorkQueueUpdate is a subset of WorkQueue used when updating queues.
type WorkQueueUpdate struct {
	Description      *string `json:"description"`
	IsPaused         *bool   `json:"is_paused"`
	ConcurrencyLimit *int64  `json:"concurrency_limit"`
	Priority         *int64  `json:"priority"`
}

// WorkQueueFilter defines filters when searching for work queues.
type WorkQueueFilter struct {
	Any []uuid.UUID `json:"any_"`
}
