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
	"sync"
)

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

type Publisher struct {
	sync.Mutex

	subs map[*Receiver]bool
}

// New creates an in-memory publisher.
func New() *Publisher {
	return &Publisher{
		subs: make(map[*Receiver]bool),
	}
}

func (p *Publisher) Publish(message Message) {
	p.Lock()
	for s := range p.subs {
		go (*s)(message)
	}
	p.Unlock()
}

func (p *Publisher) Subscribe(c context.Context, receiver Receiver) {
	p.Lock()
	p.subs[&receiver] = true
	p.Unlock()
	<-c.Done()
	p.Lock()
	delete(p.subs, &receiver)
	p.Unlock()
}
