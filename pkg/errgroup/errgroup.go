// Package errgroup provides synchronization, error propagation, and Context
// cancelation for groups of goroutines working on subtasks of a common task.
//
// This is a simpler version of golang.org/x/sync/errgroup without TryGo
// functionality. It was implemented as an exercise for me to get more
// understanding on how to propagate errors across multiple goroutines.
//
// This is great for use case where an error occurs in one goroutine within the
// group must stop all other goroutines.
package errgroup

import (
	"context"
	"sync"
)

// A group is a collection of goroutines working on subtasks that are part of
// the same overall task.
type group struct {
	ctx    context.Context
	cancel context.CancelCauseFunc
	wg     sync.WaitGroup

	// object to set err only once. Must not be copied
	errOnce sync.Once

	// error interface to return to caller
	err error
}

// WithContext returns a Group and Context derived from the parent context
// Use this function to create a Group if an error in one goroutine needs to
// propagate across goroutines in the same Group.
//
// Usage:
//
//	g := errgroup.WithContext(ctx)
func WithContext(parent context.Context) *group {
	ctx, cancel := context.WithCancelCause(parent)
	return &group{cancel: cancel, ctx: ctx}
}

// Go executes the given function in a new goroutine.
// The first call that returns non-nil error cancels the group's context,
// if the group is created using WithContext.
// Error is returned with Wait()
//
// Usage:
//
//	g := errgroup.WithContext(ctx)
//	g.Go(exampleFunc1)
//	g.Go(exampleFunc2)
//
//	if err := g.Wait(); err != nil {
//	    fmt.Println(err)
//	}
func (g *group) Go(f func(ctx context.Context) error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := f(g.ctx); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel(err)
				}
			})
		}
	}()
}

// Wait blocks until all goroutines stop execution and return the first error
// If all goroutines finish without error, it will cancel the context.
func (g *group) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel(g.err)
	}

	return g.err
}
