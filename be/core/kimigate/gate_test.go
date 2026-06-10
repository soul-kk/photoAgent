package kimigate

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestGate_AcquireRelease(t *testing.T) {
	g := New(Options{MaxConcurrent: 1, TimeoutSec: 5, QueueWaitSec: 1})
	if err := g.Acquire(context.Background()); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		done <- g.Acquire(context.Background())
	}()
	select {
	case err := <-done:
		if !errors.Is(err, ErrTooManyConcurrent) {
			t.Fatalf("expected ErrTooManyConcurrent, got %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("timeout waiting for queue full")
	}
	g.Release()
	if err := g.Acquire(context.Background()); err != nil {
		t.Fatal(err)
	}
	g.Release()
}
