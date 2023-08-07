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

import "sync"

type topic struct {
	sync.Mutex

	name string
	done chan struct{}
	subs map[*subscriber]struct{}
}

func newTopic(dest string) *topic {
	return &topic{
		name: dest,
		done: make(chan struct{}),
		subs: make(map[*subscriber]struct{}),
	}
}

func (t *topic) subscribe(s *subscriber) {
	t.Lock()
	t.subs[s] = struct{}{}
	t.Unlock()
}

func (t *topic) unsubscribe(s *subscriber) {
	t.Lock()
	delete(t.subs, s)
	t.Unlock()
}

func (t *topic) publish(m Message) {
	t.Lock()
	for s := range t.subs {
		go s.receiver(m)
	}
	t.Unlock()
}

func (t *topic) close() {
	t.Lock()
	close(t.done)
	t.Unlock()
}
