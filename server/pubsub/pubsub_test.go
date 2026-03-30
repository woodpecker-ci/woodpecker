// Copyright 2026 Woodpecker Authors
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

package pubsub_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub/memory"
)

func TestPubSub(t *testing.T) {
	// for each pubsub adapter (currently we have only one)
	t.Run("in_memory", func(t *testing.T) {
		testPubSub(t, memory.New())
	})
}

func testPubSub(t *testing.T, adapter pubsub.PubSub) {
	assert.NoError(t,
		adapter.Publish(t.Context(), pubsub.Topics{"a": {}}, pubsub.Message{ID: "1", Data: []byte(`dummy`)}),
		"expect no issue publish to a pubsub with no subscribers",
	)

	t.Run("test deduplication asumptions", func(t *testing.T) {
		treeTopicCloser, treeTopicGetMSGs := genTestSub(t, adapter, pubsub.Topics{"tree": {}})
		t.Cleanup(treeTopicCloser)
		closer, getMSGs := genTestSub(t, adapter, pubsub.Topics{"apples": {}, "tree": {}, "raspberry": {}})
		t.Cleanup(closer)
		assert.Len(t, getMSGs(), 0)

		time.Sleep(10 * time.Millisecond)
		assert.NoError(t, adapter.Publish(t.Context(), pubsub.Topics{"tree": {}, "raspberry": {}, "tails": {}}, pubsub.Message{ID: "2"}))
		assert.NoError(t, adapter.Publish(t.Context(), pubsub.Topics{"apples": {}, "raspberry": {}, "tails": {}}, pubsub.Message{ID: "3"}))
		time.Sleep(100 * time.Millisecond)

		if assert.Len(t, getMSGs(), 2) {
			assert.ElementsMatch(t, []string{"2", "3"}, messagesToIDs(getMSGs()))
		}

		assert.EqualValues(t, "2", treeTopicGetMSGs()[0].ID)
	})

	t.Run("test adapters calc for strange input", func(t *testing.T) {
		t.Run("empty topic", func(t *testing.T) {
			assert.Error(t, adapter.Subscribe(t.Context(), nil, func(pubsub.Message) {}))
			assert.Error(t, adapter.Subscribe(t.Context(), pubsub.Topics{}, func(pubsub.Message) {}))

			assert.Error(t, adapter.Publish(t.Context(), nil, pubsub.Message{}))
			assert.Error(t, adapter.Publish(t.Context(), pubsub.Topics{}, pubsub.Message{}))
		})
	})
}

func genTestSub(t *testing.T, adapter pubsub.PubSub, topics pubsub.Topics) (close func(), getMSGs func() []pubsub.Message) {
	ctx, closer := context.WithCancelCause(t.Context())
	var mu sync.Mutex
	var messages []pubsub.Message

	go func() {
		err := adapter.Subscribe(ctx, topics, func(m pubsub.Message) {
			mu.Lock()
			messages = append(messages, m)
			mu.Unlock()
		})
		assert.NoError(t, err)
	}()

	return func() { closer(nil) }, func() []pubsub.Message {
		mu.Lock()
		defer mu.Unlock()
		cp := make([]pubsub.Message, len(messages))
		copy(cp, messages)
		return cp
	}
}

func messagesToIDs(msgs []pubsub.Message) (ids []string) {
	for i := range msgs {
		ids = append(ids, msgs[i].ID)
	}
	return ids
}
