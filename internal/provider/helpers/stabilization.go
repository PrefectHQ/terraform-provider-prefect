package helpers

import (
	"context"
	"fmt"
	"time"

	"github.com/avast/retry-go/v4"
)

const (
	// DefaultStabilizationAttempts is the default number of retry attempts for resource stabilization.
	DefaultStabilizationAttempts = 10
	// DefaultStabilizationDelay is the default delay between retry attempts.
	DefaultStabilizationDelay = 500 * time.Millisecond
)

// WaitForResourceStabilization is a generic helper that retries fetching a resource
// until a comparison function indicates the state is stable.
//
// This is useful for handling eventual consistency in the Prefect API, where a resource
// may be created or updated but take some time for all fields to reach their final values.
//
// Type parameter T is the resource type (e.g., *api.ServiceAccount, *api.Workspace).
//
// Parameters:
//   - ctx: Context for the operation
//   - fetchFunc: Function that fetches the current resource state from the API
//   - isStableFunc: Function that returns nil if the state is stable, or an error describing why it's not
//
// The function will retry up to DefaultStabilizationAttempts times with DefaultStabilizationDelay
// between attempts. If all retries are exhausted, it returns the last fetched state anyway,
// allowing Terraform to detect any remaining drift.
//
// Example usage:
//
//	serviceAccount, err := WaitForResourceStabilization(
//	    ctx,
//	    func(ctx context.Context) (*api.ServiceAccount, error) {
//	        return client.Get(ctx, id)
//	    },
//	    func(sa *api.ServiceAccount) error {
//	        if sa.Name != expectedName {
//	            return fmt.Errorf("name mismatch: got %q, want %q", sa.Name, expectedName)
//	        }
//	        return nil
//	    },
//	)
//
//nolint:ireturn // Generic functions must return type parameter
func WaitForResourceStabilization[T any](
	ctx context.Context,
	fetchFunc func(context.Context) (T, error),
	isStableFunc func(T) error,
) (T, error) {
	var resource T
	var fetchErr error

	retryErr := retry.Do(
		func() error {
			var err error
			resource, err = fetchFunc(ctx)
			if err != nil {
				fetchErr = err

				return fmt.Errorf("failed to fetch resource: %w", err)
			}

			// Check if state is stable
			if err := isStableFunc(resource); err != nil {
				return err
			}

			return nil
		},
		retry.Attempts(DefaultStabilizationAttempts),
		retry.Delay(DefaultStabilizationDelay),
		retry.LastErrorOnly(true),
	)

	if retryErr != nil {
		// Even if we hit max retries, return the last resource state if we have one
		// This allows Terraform to proceed and detect any remaining drift
		if fetchErr == nil {
			return resource, nil
		}

		var zero T

		return zero, fmt.Errorf("failed to stabilize resource state: %w", retryErr)
	}

	return resource, nil
}

// WaitForResourceStabilizationByComparison is a variant that compares consecutive reads
// to detect when state stops changing. This is useful when you don't know the expected
// final values, but want to wait until the API stops transforming the data.
//
// Type parameter T is the resource type (e.g., *api.Automation).
//
// Parameters:
//   - ctx: Context for the operation
//   - fetchFunc: Function that fetches the current resource state from the API
//   - compareFunc: Function that returns true if two consecutive states are equal
//
// The function will retry up to DefaultStabilizationAttempts times with DefaultStabilizationDelay
// between attempts. If all retries are exhausted, it returns the last fetched state anyway.
//
// Example usage for automation's match_related field:
//
//	automation, err := WaitForResourceStabilizationByComparison(
//	    ctx,
//	    func(ctx context.Context) (*api.Automation, error) {
//	        return client.Get(ctx, id)
//	    },
//	    func(prev, curr *api.Automation) bool {
//	        prevJSON, _ := json.Marshal(prev.Trigger.MatchRelated)
//	        currJSON, _ := json.Marshal(curr.Trigger.MatchRelated)
//	        return bytes.Equal(prevJSON, currJSON)
//	    },
//	)
//
//nolint:ireturn // Generic functions must return type parameter
func WaitForResourceStabilizationByComparison[T any](
	ctx context.Context,
	fetchFunc func(context.Context) (T, error),
	compareFunc func(prev, curr T) bool,
) (T, error) {
	var resource T
	var lastResource T
	var fetchErr error
	isFirstAttempt := true

	retryErr := retry.Do(
		func() error {
			var err error
			resource, err = fetchFunc(ctx)
			if err != nil {
				fetchErr = err

				return fmt.Errorf("failed to fetch resource: %w", err)
			}

			// If this is not the first attempt, check if state has stabilized
			if !isFirstAttempt && compareFunc(lastResource, resource) {
				// State has stabilized
				return nil
			}

			// State is still changing (or this is the first attempt), save current state and retry
			lastResource = resource
			isFirstAttempt = false

			return fmt.Errorf("resource state still changing")
		},
		retry.Attempts(DefaultStabilizationAttempts),
		retry.Delay(DefaultStabilizationDelay),
		retry.LastErrorOnly(true),
	)

	if retryErr != nil {
		// Even if we hit max retries, return the last resource state if we have one
		if fetchErr == nil {
			return resource, nil
		}

		var zero T

		return zero, fmt.Errorf("failed to stabilize resource state: %w", retryErr)
	}

	return resource, nil
}
