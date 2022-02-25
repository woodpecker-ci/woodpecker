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

package exec

import (
	"context"
	"fmt"
	"os"
	"os/signal"
)

// registerInterruptSignal returns a copy of parent context that is canceled when an os interrupt signal is received.
// it will print this information to stdout
func registerInterruptSignal(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		defer signal.Stop(c)

		select {
		case <-ctx.Done():
		case <-c:
			fmt.Println("ctrl+c received, terminating process")
			cancel()
		}
	}()

	return ctx
}
