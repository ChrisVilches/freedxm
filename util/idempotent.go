package util

import (
	"context"
	"sync"
)

type IdempotentRunner struct {
	fn func(context.Context)
	mu sync.Mutex
}

func NewIdempotentRunner(fn func(context.Context)) IdempotentRunner {
	return IdempotentRunner{fn: fn}
}

func (runner *IdempotentRunner) Run(ctx context.Context) {
	if !runner.mu.TryLock() {
		return
	}

	defer runner.mu.Unlock()

	runner.fn(ctx)
}
