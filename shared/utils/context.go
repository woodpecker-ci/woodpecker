package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// Returns a copy of parent context that is canceled when
// an os interrupt signal is received.
func WithContextSigtermCallback(ctx context.Context, f func()) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		recivedSignal := make(chan os.Signal, 1)
		signal.Notify(recivedSignal, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(recivedSignal)

		select {
		case <-ctx.Done():
		case <-recivedSignal:
			cancel()
			if f != nil {
				f()
			}
		}
	}()

	return ctx
}
