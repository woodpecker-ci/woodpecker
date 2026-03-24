// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memory

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub/types"
)

func TestPubsub(t *testing.T) {
	var (
		wg sync.WaitGroup

		testMessage = types.Message{
			Data: []byte("test"),
		}
	)

	ctx, cancel := context.WithCancelCause(
		t.Context(),
	)

	broker := New()
	go func() {
		broker.Subscribe(ctx, func(message types.Message) { assert.Equal(t, testMessage, message); wg.Done() })
	}()
	go func() {
		broker.Subscribe(ctx, func(_ types.Message) { wg.Done() })
	}()

	<-time.After(500 * time.Millisecond)

	wg.Add(2)
	go func() {
		broker.Publish(testMessage)
	}()

	wg.Wait()
	cancel(nil)
}
