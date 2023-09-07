package xtime

import (
	"context"
	"time"
)

// -----------------------------------------------------------------------------------------

// SleepContext pauses the current goroutine for at least the duration d. A negative or zero duration causes Sleep to return immediately.
// Sleep may be canceled by context.Context.
func SleepContext(ctx context.Context, d time.Duration) (err error) {
	t := time.NewTimer(d)
	select {
	case <-t.C:
	case <-ctx.Done():
		err = context.Canceled
	}
	t.Stop()
	return
}

// -----------------------------------------------------------------------------------------
