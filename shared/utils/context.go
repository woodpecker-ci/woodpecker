package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

func WithContextSigterm(ctx context.Context) context.Context {
	return WithContextSigtermCallback(ctx, nil)
}

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
		case signal := <-recivedSignal:
			cancel()
			if f != nil {
				f()
			}
			log.Warn().Str(
				"SIG", signal.String(),
			).Msg("Received termination signal")
			break
		}
	}()

	return ctx
}
