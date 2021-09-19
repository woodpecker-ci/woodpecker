// Copyright 2017 Drone.IO Inc.
//
// This file is licensed under the terms of the MIT license.
// For a copy, see https://opensource.org/licenses/MIT.

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WithContext returns a copy of parent context whose Done channel is closed
// when an os interrupt signal is received.
func WithContext(ctx context.Context) context.Context {
	return WithContextFunc(ctx, func() {
		println("interrupt received, terminating process")
	})
}

// WithContextFunc returns a copy of parent context that is cancelled when
// an os interrupt signal is received. The callback function f is invoked
// before cancellation.
func WithContextFunc(ctx context.Context, f func()) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(c)

		select {
		case <-ctx.Done():
		case <-c:
			f()
			cancel()
		}
	}()

	return ctx
}
