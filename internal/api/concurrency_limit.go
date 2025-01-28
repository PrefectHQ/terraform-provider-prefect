package api

import (
	"context"

	"github.com/google/uuid"
)

// ConcurrencyLimitsClient is a client for working with concurrency limits.
type ConcurrencyLimitsClient interface {
	Create(ctx context.Context, concurrencyLimit ConcurrencyLimitCreate) (*ConcurrencyLimit, error)
	Get(ctx context.Context, concurrencyLimitID uuid.UUID) (*ConcurrencyLimit, error)
	GetByTag(ctx context.Context, tag string) (*ConcurrencyLimit, error)
	Delete(ctx context.Context, concurrencyLimitID uuid.UUID) error
	DeleteByTag(ctx context.Context, tag string) error
}

// ConcurrencyLimit is a representation of a concurrency limit.
type ConcurrencyLimit struct {
	BaseModel
	Tag              string `json:"tag"`
	ConcurrencyLimit int    `json:"concurrency_limit"`
}

// ConcurrencyLimitCreate is a subset of ConcurrencyLimit used when creating concurrency limits.
type ConcurrencyLimitCreate struct {
	Tag              string `json:"tag"`
	ConcurrencyLimit int    `json:"concurrency_limit"`
}
