package kubectl

import (
	"sync"
)

type WaitOnce struct {
	mutex     sync.Mutex
	waiting   []chan struct{}
	err       error
	completed bool
}

func (wait *WaitOnce) Completed() bool {
	return wait.completed
}

func (wait *WaitOnce) MarkComplete(err error) {
	wait.mutex.Lock()
	if !wait.completed {
		wait.err = err
		wait.completed = true
		for _, waiter := range wait.waiting {
			waiter <- struct{}{}
		}
	}
	wait.mutex.Unlock()
}

func (wait *WaitOnce) Wait() error {
	// would only achieve lock if has been completed.
	var waiter chan struct{}
	wait.mutex.Lock()
	if !wait.completed {
		waiter = make(chan struct{})
		wait.waiting = append(wait.waiting, waiter)
	}
	wait.mutex.Unlock()

	if waiter != nil {
		<-waiter
	}

	return wait.err
}
