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

package memory

import (
	"context"
	"sync"

	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub/types"
)

type Publisher struct {
	sync.Mutex

	subs map[*types.Receiver]struct{}
}

// New creates an in-memory publisher.
func New() *Publisher {
	return &Publisher{
		subs: make(map[*types.Receiver]struct{}),
	}
}

func (p *Publisher) Publish(message types.Message) {
	p.Lock()
	for s := range p.subs {
		go (*s)(message)
	}
	p.Unlock()
}

func (p *Publisher) Subscribe(c context.Context, receiver types.Receiver) {
	p.Lock()
	p.subs[&receiver] = struct{}{}
	p.Unlock()
	<-c.Done()
	p.Lock()
	delete(p.subs, &receiver)
	p.Unlock()
}
