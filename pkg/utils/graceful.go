package utils

import (
	"errors"
	"sync"
)

// GracefulContext - a context carries channel factory for cancelation
type GracefulContext interface {
	// Done - blocks until cancelled
	Done() <-chan struct{}

	// RunAsChild - runs handler within child context
	RunAsChild(callback func(GracefulContext)) GracefulRunner

	// Fail - cancels run from the inside and propagates cancel to all children
	// Does not await the cancellation (obviously)
	Fail(err error)
}

// GracefulRunner - outter API for controlling gracefully run handlers
type GracefulRunner interface {
	// Wait - blocks until finishes execution
	Wait() (bool, error)

	// Cancel - notifies handler to cancel the execution and awaits graceful return (clean up)
	Cancel()
}

// RunWithGracefulCancel - runs callback as a go routine and returns cancel routine
// This is inspired by context but with the key difference that the cancel function waits until
// the handler finishes all the cleanup
// @see https://blog.golang.org/context
func RunWithGracefulCancel(callback func(GracefulContext)) GracefulRunner {
	ctx := newGracefulCtx()
	ctx.wg.Add(1)

	go func() {
		callback(ctx)
		ctx.wg.Done()
	}()

	return newGracefulRunner(ctx)
}

// -----------------------------

type gracefulRunner struct {
	ctx *gracefulCtx
}

func newGracefulRunner(ctx *gracefulCtx) *gracefulRunner {
	return &gracefulRunner{ctx}
}

func (runner *gracefulRunner) Wait() (bool, error) {
	runner.ctx.wg.Wait()
	return runner.ctx.err == errCancel, runner.ctx.err
}

func (runner *gracefulRunner) Cancel() {
	runner.ctx.Fail(errCancel)
	runner.ctx.wg.Wait()
}

// -----------------------------

type gracefulCtx struct {
	cancelC         chan struct{}
	wg              sync.WaitGroup
	mutex           sync.Mutex
	hasBeenCanceled bool
	err             error
}

func newGracefulCtx() *gracefulCtx {
	return &gracefulCtx{
		cancelC:         make(chan struct{}),
		hasBeenCanceled: false,
	}
}

func (c *gracefulCtx) Done() <-chan struct{} {
	return c.cancelC
}

func (c *gracefulCtx) RunAsChild(callback func(GracefulContext)) GracefulRunner {
	// Parent should wait for 1 more child
	c.wg.Add(1)

	// Do not even start if context has been already cancelled
	c.mutex.Lock()
	if c.hasBeenCanceled {
		c.wg.Done()
		c.mutex.Unlock()
		return newCancelledGracefulRunner()
	}
	c.mutex.Unlock()

	// Create channel for cancel callback
	// Note: this is necessary so that we don't leak if child finishes first
	cancelC := make(chan struct{}, 1)

	// Start child runner
	runner := RunWithGracefulCancel(func(childCtx GracefulContext) {
		callback(childCtx)

		// Notify parent that child is done
		c.wg.Done()

		// Unblock cancel callback
		close(cancelC)
	})

	// Cancel child when parent gets cancelled
	go func() {
		select {
		case <-c.Done():
			runner.Cancel()
		case <-cancelC:
			return
		}
	}()

	return runner
}

func (c *gracefulCtx) Fail(err error) {
	c.mutex.Lock()
	if c.hasBeenCanceled {
		c.mutex.Unlock()
		return
	}

	c.hasBeenCanceled = true
	c.err = err
	close(c.cancelC)
	c.mutex.Unlock()
}

// -----------------------------

type cancelledGracefulRunner struct{}

func newCancelledGracefulRunner() *cancelledGracefulRunner {
	return &cancelledGracefulRunner{}
}

func (*cancelledGracefulRunner) Wait() (bool, error) {
	return true, errCancel
}

func (*cancelledGracefulRunner) Cancel() {}

var errCancel error = errors.New("cancelled execution")
