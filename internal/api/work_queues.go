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
	Name             string  `json:"name"`
	Description      *string `json:"description"`
	IsPaused         bool    `json:"is_paused"`
	ConcurrencyLimit *int64  `json:"concurrency_limit"`
	Priority         *int64  `json:"priority"`
	// not yet required by the api
	WorkPoolID *uuid.UUID `json:"work_pool_id"`
	LastPolled *string    `json:"last_polled"`
	Status     *string    `json:"status"`
	WorkPool   *WorkPool  `json:"work_pool"`
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
	Name             string  `json:"name"`
	Description      *string `json:"description"`
	IsPaused         bool    `json:"is_paused"`
	ConcurrencyLimit *int64  `json:"concurrency_limit"`
	Priority         *int64  `json:"priority"`
	LastPolled       string  `json:"last_polled"`
	Status           string  `json:"status"`
}

// WorkQueueFilter defines filters when searching for work queues.
type WorkQueueFilter struct {
	Any []uuid.UUID `json:"any_"`
}
