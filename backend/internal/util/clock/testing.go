package clock

import (
	"context"
	"sync/atomic"
	"time"
)

// WithFixedTime returns a context whose clock always returns t.
func WithFixedTime(ctx context.Context, t time.Time) context.Context {
	return WithGenerator(ctx, func() time.Time { return t })
}

// WithSequentialTimes returns a context whose clock yields times in order.
// It panics when all times have been consumed.
func WithSequentialTimes(ctx context.Context, times ...time.Time) context.Context {
	var idx atomic.Int64
	return WithGenerator(ctx, func() time.Time {
		i := int(idx.Add(1) - 1)
		if i >= len(times) {
			panic("clock.WithSequentialTimes: all times exhausted")
		}
		return times[i]
	})
}
