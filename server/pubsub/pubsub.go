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

// Package pubsub implements a publish-subscriber messaging system.
package pubsub

import (
	"context"
	"errors"
)

// ErrNotFound is returned when the named topic does not exist.
var ErrNotFound = errors.New("topic not found")

// Message defines a published message.
type Message struct {
	// ID identifies this message.
	ID string `json:"id,omitempty"`

	// Data is the actual data in the entry.
	Data []byte `json:"data"`

	// Labels represents the key-value pairs the entry is labeled with.
	Labels map[string]string `json:"labels,omitempty"`
}

// Receiver receives published messages.
type Receiver func(Message)

// Publisher defines a mechanism for communicating messages from a group
// of senders, called publishers, to a group of consumers.
type Publisher interface {
	// Create creates the named topic.
	Create(c context.Context, topic string) error

	// Publish publishes the message.
	Publish(c context.Context, topic string, message Message) error

	// Subscribe subscribes to the topic. The Receiver function is a callback
	// function that receives published messages.
	Subscribe(c context.Context, topic string, receiver Receiver) error

	// Remove removes the named topic.
	Remove(c context.Context, topic string) error
}
