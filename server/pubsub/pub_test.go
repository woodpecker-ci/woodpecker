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

package pubsub

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPubsub(t *testing.T) {
	var (
		wg sync.WaitGroup

		testTopic   = "test"
		testMessage = Message{
			Data: []byte("test"),
		}
	)

	ctx, cancel := context.WithCancelCause(
		context.Background(),
	)

	broker := New()
	assert.NoError(t, broker.Create(ctx, testTopic))
	go func() {
		assert.NoError(t, broker.Subscribe(ctx, testTopic, func(message Message) { wg.Done() }))
	}()
	go func() {
		assert.NoError(t, broker.Subscribe(ctx, testTopic, func(message Message) { wg.Done() }))
	}()

	<-time.After(500 * time.Millisecond)

	if _, ok := broker.(*publisher).topics[testTopic]; !ok {
		t.Errorf("Expect topic registered with publisher")
	}

	wg.Add(2)
	go func() {
		assert.NoError(t, broker.Publish(ctx, testTopic, testMessage))
	}()

	wg.Wait()
	cancel(nil)
}

func TestPublishNotFound(t *testing.T) {
	var (
		testTopic   = "test"
		testMessage = Message{
			Data: []byte("test"),
		}
	)
	broker := New()
	err := broker.Publish(context.Background(), testTopic, testMessage)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expect Not Found error when topic does not exist")
	}
}

func TestSubscribeNotFound(t *testing.T) {
	var (
		testTopic    = "test"
		testCallback = func(message Message) {}
	)
	broker := New()
	err := broker.Subscribe(context.Background(), testTopic, testCallback)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("Expect Not Found error when topic does not exist")
	}
}

func TestSubscriptionClosed(t *testing.T) {
	var (
		wg sync.WaitGroup

		testTopic    = "test"
		testCallback = func(Message) {}
	)

	broker := New()
	assert.NoError(t, broker.Create(context.Background(), testTopic))
	go func() {
		assert.NoError(t, broker.Subscribe(context.Background(), testTopic, testCallback))
		wg.Done()
	}()

	<-time.After(500 * time.Millisecond)

	if _, ok := broker.(*publisher).topics[testTopic]; !ok {
		t.Errorf("Expect topic registered with publisher")
	}

	wg.Add(1)
	assert.NoError(t, broker.Remove(context.Background(), testTopic))
	wg.Wait()

	if _, ok := broker.(*publisher).topics[testTopic]; ok {
		t.Errorf("Expect topic removed from publisher")
	}
}
