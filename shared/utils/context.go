// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

// Returns a copy of parent context that is canceled when
// an os interrupt signal is received.
func WithContextSigtermCallback(ctx context.Context, f func()) context.Context {
	ctx, cancel := context.WithCancelCause(ctx)
	go func() {
		receivedSignal := make(chan os.Signal, 1)
		signal.Notify(receivedSignal, syscall.SIGINT, syscall.SIGTERM)
		defer signal.Stop(receivedSignal)

		select {
		case <-ctx.Done():
		case <-receivedSignal:
			if f != nil {
				f()
			}
			cancel(fmt.Errorf("received signal: %v", receivedSignal))
		}
	}()

	return ctx
}
