package cmdserve

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestEvery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var runs atomic.Int32
	done := make(chan struct{})
	go func() {
		every(ctx, time.Millisecond, func() { runs.Add(1) })
		close(done)
	}()

	deadline := time.After(5 * time.Second)
	for runs.Load() < 2 {
		select {
		case <-deadline:
			t.Fatal("the task did not run twice")
		case <-time.After(time.Millisecond):
		}
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("every did not stop with its context")
	}
}
