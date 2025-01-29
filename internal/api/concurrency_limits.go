package api

import (
	"context"
)

// ConcurrencyLimitsClient is a client for working with concurrency limits.
type ConcurrencyLimitsClient interface {
	Create(ctx context.Context, concurrencyLimit ConcurrencyLimitCreate) (*ConcurrencyLimit, error)
	Read(ctx context.Context, concurrencyLimitID string) (*ConcurrencyLimit, error)
	Delete(ctx context.Context, concurrencyLimitID string) error
}

// ConcurrencyLimit is a representation of a concurrency limit.
type ConcurrencyLimit struct {
	BaseModel
	Tag              string `json:"tag"`
	ConcurrencyLimit int64  `json:"concurrency_limit"`
}

// ConcurrencyLimitCreate is a subset of ConcurrencyLimit used when creating concurrency limits.
type ConcurrencyLimitCreate struct {
	Tag              string `json:"tag"`
	ConcurrencyLimit int64  `json:"concurrency_limit"`
}
