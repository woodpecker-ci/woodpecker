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

package pubsub

import (
	"context"
	"errors"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// PubSub provider interface, used to signal pipeline state changes to WebUI.
type PubSub interface {
	// Publish pushes a state change to all subscribers,
	// that have at least subscribed to one of the topics we publish it under.
	Publish(context.Context, Topics, Message) error
	// Subscribe gets all state changes that match the same topic.
	// If multiple topics are subscribed, and a message also match multiple,
	// the implementation takes care of deduplication.
	Subscribe(context.Context, Topics, Receiver) error
}

// Message defines a published message.
type Message struct {
	// ID identifies this message.
	ID string `json:"id,omitempty"`

	// Data is the actual data in the entry.
	Data []byte `json:"data"`
}

// Receiver receives published messages.
type Receiver func(Message)

// Topics are key-value pairs, messages are filtered upon
// the the key is the base-key and the value to the sub-key.
type Topics map[string]struct{}

func GetRepoTopic(r *model.Repo) string {
	return fmt.Sprintf("repo.id.%d", r.ID)
}

const PublicTopic = "public"

var ErrNoTopic = errors.New("no topic specified")
