package util

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestIdempotentRunner_RunsFunction(t *testing.T) {
	called := false
	runner := NewIdempotentRunner(func(_ context.Context) {
		called = true
	})

	runner.Run(context.Background())

	if !called {
		t.Errorf("expected function to be called, but it was not")
	}
}

func TestIdempotentRunner_Idempotency(t *testing.T) {
	var count int
	var wg sync.WaitGroup
	runner := NewIdempotentRunner(func(_ context.Context) {
		time.Sleep(100 * time.Millisecond)
		count++
	})

	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			runner.Run(context.Background())
		}()
	}

	wg.Wait()

	if count != 1 {
		t.Errorf("expected function to be called once, but was called %d times", count)
	}
}
