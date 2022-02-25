package kubectl

import (
	"sync"

	"golang.org/x/net/context"
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

type ActionContext struct {
	// internal
	cancel      context.CancelFunc // action context cancel.
	ctx         context.Context    // action context
	startWaiter WaitOnce           // a startup waiter.
	isRunning   bool               // True if completed.
	OnStop      func(err error)    // Called when stopped.
	mutex       sync.Mutex         // the lock action mutex
}

// Only available if action has started. The action context.
func (action *ActionContext) Context() context.Context {
	return action.ctx
}

// True if the action is running.
func (action *ActionContext) IsRunning() bool {
	return action.isRunning
}

// Stop the action. Returns true if stopped.
func (action *ActionContext) Stop(err error) bool {
	action.mutex.Lock()
	action.cancel()
	if !action.isRunning {
		action.startWaiter.MarkComplete(nil)
		action.mutex.Unlock()
		return false
	}
	action.startWaiter.MarkComplete(err)
	action.isRunning = false
	if action.OnStop != nil {
		action.OnStop(err)
	}
	action.mutex.Unlock()
	return true
}

func (action *ActionContext) Start(
	ctx context.Context,
	invoke func(),
) {
	action.mutex.Lock()
	action.isRunning = true
	action.ctx, action.cancel = context.WithCancel(ctx)
	action.mutex.Unlock()

	go func() {
		<-action.ctx.Done()
		action.Stop(action.ctx.Err())
	}()

	go func() {
		invoke()
		action.Stop(nil)
	}()
}

func (action *ActionContext) MarkActionStarted() {
	action.startWaiter.MarkComplete(nil)
}

func (action *ActionContext) WaitForActionStarted() error {
	return action.startWaiter.Wait()
}

func (action *ActionContext) Wait() error {
	<-action.ctx.Done()
	return action.ctx.Err()
}
