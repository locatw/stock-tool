// Package idp provides context-based UUID generation for domain entities.
// In production, NewV7 falls back to uuid.NewV7(). In tests, inject a
// deterministic generator via WithGenerator, WithFixedID, or
// WithSequentialIDs.
package idp

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct{}

// Generator produces a UUID. Implementations must be safe for concurrent
// use only if the caller shares the context across goroutines.
type Generator func() uuid.UUID

// NewV7 returns a UUIDv7 from the generator stored in ctx, or falls back
// to uuid.NewV7().
func NewV7(ctx context.Context) uuid.UUID {
	if gen, ok := ctx.Value(ctxKey{}).(Generator); ok {
		return gen()
	}
	return uuid.Must(uuid.NewV7())
}

// WithGenerator returns a child context that uses gen for all NewV7 calls.
func WithGenerator(ctx context.Context, gen Generator) context.Context {
	return context.WithValue(ctx, ctxKey{}, gen)
}
