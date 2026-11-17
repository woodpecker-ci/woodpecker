// Copyright 2026 Woodpecker Authors
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

package scheduler

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
)

type capturePubSub struct {
	pubsub.PubSub
	msg pubsub.Message
}

func (c *capturePubSub) Publish(_ context.Context, _ pubsub.Topics, m pubsub.Message) error {
	c.msg = m
	return nil
}

// The web UI reads the deprecated pipeline fields from the event stream.
func TestPublishPipelineEventSendsAPIPipeline(t *testing.T) {
	ps := new(capturePubSub)
	s := NewScheduler(t.Context(), nil, nil, ps)

	require.NoError(t, s.PublishPipelineEvent(t.Context(), &model.Repo{ID: 1}, &model.Pipeline{
		Number: 2,
		Event:  model.EventPush,
		Commit: &model.Commit{SHA: "abc123", Message: "fix the thing"},
	}))

	var event struct {
		Pipeline map[string]any `json:"pipeline"`
	}
	require.NoError(t, json.Unmarshal(ps.msg.Data, &event))
	assert.Equal(t, "abc123", event.Pipeline["commit"])
	assert.Equal(t, "fix the thing", event.Pipeline["message"])
}
