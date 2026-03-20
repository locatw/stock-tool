// Package clock provides context-based time generation for domain entities.
// In production, Now falls back to time.Now(). In tests, inject a
// deterministic generator via WithGenerator, WithFixedTime, or
// WithSequentialTimes.
package clock

import (
	"context"
	"time"
)

type ctxKey struct{}

// Generator returns the current time. Implementations must be safe for
// concurrent use only if the caller shares the context across goroutines.
type Generator func() time.Time

// Now returns the current time from the generator stored in ctx, or falls
// back to time.Now().
func Now(ctx context.Context) time.Time {
	if c, ok := ctx.Value(ctxKey{}).(Generator); ok {
		return c()
	}
	return time.Now()
}

// WithGenerator returns a child context that uses g for all Now calls.
func WithGenerator(ctx context.Context, g Generator) context.Context {
	return context.WithValue(ctx, ctxKey{}, g)
}
