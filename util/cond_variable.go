package util

import (
	"sync"
)

type CondVar struct {
	cond  *sync.Cond
	mutex sync.Locker
}

func NewCondVar(m sync.Locker) *CondVar {
	return &CondVar{
		cond:  sync.NewCond(m),
		mutex: m,
	}
}

func (cv *CondVar) WaitUntil(predicate func() bool) {
	cv.mutex.Lock()
	for !predicate() {
		// TODO: Specially this part is suspicious. Wait (docs)
		// mentions something about
		// unlocking (?) the mutex so why do we need another unlock at the end?
		// also, read the documentation in detail about this
		// part of waiting in a loop.
		cv.cond.Wait()
	}
	cv.mutex.Unlock()
}

func (cv *CondVar) Signal() {
	cv.cond.Signal()
}

func (cv *CondVar) Broadcast() {
	cv.cond.Broadcast()
}
