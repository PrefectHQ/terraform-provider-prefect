package api

import (
	"context"
)

// TaskRunConcurrencyLimitsClient is a client for working with task run concurrency limits (in the api this is named "concurrency_limit").
type TaskRunConcurrencyLimitsClient interface {
	Create(ctx context.Context, taskRunConcurrencyLimit TaskRunConcurrencyLimitCreate) (*TaskRunConcurrencyLimit, error)
	Read(ctx context.Context, taskRunConcurrencyLimitID string) (*TaskRunConcurrencyLimit, error)
	Delete(ctx context.Context, taskRunConcurrencyLimitID string) error
}

// TaskRunConcurrencyLimit is a representation of a task run concurrency limit.
type TaskRunConcurrencyLimit struct {
	BaseModel
	Tag              string `json:"tag"`
	ConcurrencyLimit int64  `json:"concurrency_limit"`
}

// TaskRunConcurrencyLimitCreate is a subset of TaskRunConcurrencyLimit used when creating task run concurrency limits.
type TaskRunConcurrencyLimitCreate struct {
	Tag              string `json:"tag"`
	ConcurrencyLimit int64  `json:"concurrency_limit"`
}
