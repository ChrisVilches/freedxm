package util

import (
	"sync/atomic"
)

type IdempotentRunner struct {
	fn        func()
	isRunning atomic.Bool
}

func NewIdempotentRunner(fn func()) IdempotentRunner {
	return IdempotentRunner{fn: fn}
}

func (runner *IdempotentRunner) Run() {
	if !runner.isRunning.CompareAndSwap(false, true) {
		return
	}

	runner.fn()

	runner.isRunning.Store(false)
}
