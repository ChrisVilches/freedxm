package util

import (
	"context"
	"sync/atomic"
)

type IdempotentRunner struct {
	fn        func(context.Context)
	isRunning atomic.Bool
}

func NewIdempotentRunner(fn func(context.Context)) IdempotentRunner {
	return IdempotentRunner{fn: fn}
}

func (runner *IdempotentRunner) Run(ctx context.Context) {
	if !runner.isRunning.CompareAndSwap(false, true) {
		return
	}

	runner.fn(ctx)

	runner.isRunning.Store(false)
}
