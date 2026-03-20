package idp

import (
	"context"
	"sync/atomic"

	"github.com/google/uuid"
)

// WithFixedID returns a context whose generator always returns id.
func WithFixedID(ctx context.Context, id uuid.UUID) context.Context {
	return WithGenerator(ctx, func() uuid.UUID { return id })
}

// WithSequentialIDs returns a context whose generator yields ids in order.
// It panics when all IDs have been consumed.
func WithSequentialIDs(ctx context.Context, ids ...uuid.UUID) context.Context {
	var idx atomic.Int64
	return WithGenerator(ctx, func() uuid.UUID {
		i := int(idx.Add(1) - 1)
		if i >= len(ids) {
			panic("idp.WithSequentialIDs: all IDs exhausted")
		}
		return ids[i]
	})
}
