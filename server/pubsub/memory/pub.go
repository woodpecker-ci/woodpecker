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
	"fmt"
	"slices"
	"sync"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v3/server/pubsub"
)

type publisher struct {
	sync.RWMutex

	subs map[*pubsub.Receiver][]string
}

// New creates an in-memory publisher.
func New() pubsub.PubSub {
	return &publisher{
		subs: make(map[*pubsub.Receiver][]string),
	}
}

func (p *publisher) Publish(_ context.Context, topics pubsub.Topics, message pubsub.Message) error {
	if len(topics) == 0 {
		return fmt.Errorf("%w: specify at least one", pubsub.ErrNoTopic)
	}

	p.RLock()
	defer p.RUnlock()

	for s, tl := range p.subs {
		// callback is from outside so just make sure it still exists
		if s == nil || *s == nil {
			log.Error().Msg("found nil callback func in subscribers!")
			continue
		}

		for t := range topics {
			if slices.Contains(tl, t) {
				go (*s)(message)
				break
			}
		}
	}

	return nil
}

func (p *publisher) Subscribe(c context.Context, topics pubsub.Topics, receiver pubsub.Receiver) error {
	if len(topics) == 0 {
		return fmt.Errorf("%w: subscribe to at least one", pubsub.ErrNoTopic)
	}

	var tl []string
	for k := range topics {
		tl = append(tl, k)
	}

	defer func() {
		p.Lock()
		delete(p.subs, &receiver)
		p.Unlock()
	}()

	p.Lock()
	p.subs[&receiver] = tl
	p.Unlock()

	<-c.Done()
	return nil
}
