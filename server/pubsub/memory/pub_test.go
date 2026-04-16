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

	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
)

// TestPubsubConcurrentCancel verifies no panic occurs when publish and cancel race.
func TestPubsubConcurrentCancel(t *testing.T) {
	testTopic := map[string]struct{}{"test": {}}
	broker := New()

	for range 50 {
		ctx, cancel := context.WithCancelCause(t.Context())
		ch := make(chan []byte, 1)

		go func() {
			_ = broker.Subscribe(ctx, testTopic, func(m pubsub.Message) {
				select {
				case <-ctx.Done():
				case ch <- m.Data:
				}
			})
		}()

		<-time.After(10 * time.Millisecond)

		var wg sync.WaitGroup
		for range 10 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = broker.Publish(ctx, testTopic, pubsub.Message{Data: []byte("x")})
			}()
		}
		cancel(nil)
		wg.Wait()
	}
}

func TestPubsub(t *testing.T) {
	var (
		wg sync.WaitGroup

		testTopic = map[string]struct{}{"test": {}}

		testMessage = pubsub.Message{
			Data: []byte("test"),
		}
	)

	ctx, cancel := context.WithCancelCause(
		t.Context(),
	)
	broker := New()

	assert.Error(t, broker.Subscribe(ctx, nil, func(pubsub.Message) {}))
	go func() {
		assert.NoError(t, broker.Subscribe(ctx, testTopic, func(message pubsub.Message) { assert.Equal(t, testMessage, message); wg.Done() }))
	}()
	go func() {
		assert.NoError(t, broker.Subscribe(ctx, testTopic, func(pubsub.Message) { wg.Done() }))
	}()

	<-time.After(500 * time.Millisecond)

	wg.Add(2)
	go func() {
		assert.NoError(t, broker.Publish(ctx, testTopic, testMessage))
	}()

	wg.Wait()
	cancel(nil)
}
