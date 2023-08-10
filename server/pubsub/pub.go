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

type subscriber struct {
	receiver Receiver
}

type publisher struct {
	sync.Mutex

	topics map[string]*topic
}

// New creates an in-memory publisher.
func New() Publisher {
	return &publisher{
		topics: make(map[string]*topic),
	}
}

func (p *publisher) Create(_ context.Context, dest string) error {
	p.Lock()
	_, ok := p.topics[dest]
	if !ok {
		t := newTopic(dest)
		p.topics[dest] = t
	}
	p.Unlock()
	return nil
}

func (p *publisher) Publish(_ context.Context, dest string, message Message) error {
	p.Lock()
	t, ok := p.topics[dest]
	p.Unlock()
	if !ok {
		return ErrNotFound
	}
	t.publish(message)
	return nil
}

func (p *publisher) Subscribe(c context.Context, dest string, receiver Receiver) error {
	p.Lock()
	t, ok := p.topics[dest]
	p.Unlock()
	if !ok {
		return ErrNotFound
	}
	s := &subscriber{
		receiver: receiver,
	}
	t.subscribe(s)
	select {
	case <-c.Done():
	case <-t.done:
	}
	t.unsubscribe(s)
	return nil
}

func (p *publisher) Remove(_ context.Context, dest string) error {
	p.Lock()
	t, ok := p.topics[dest]
	if ok {
		delete(p.topics, dest)
		t.close()
	}
	p.Unlock()
	return nil
}
