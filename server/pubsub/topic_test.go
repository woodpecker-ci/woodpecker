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
	"testing"
	"time"
)

func TestTopicSubscribe(t *testing.T) {
	sub := new(subscriber)
	top := newTopic("foo")
	top.subscribe(sub)
	if _, ok := top.subs[sub]; !ok {
		t.Errorf("Expect subscription registered with topic on subscribe")
	}
}

func TestTopicUnsubscribe(t *testing.T) {
	sub := new(subscriber)
	top := newTopic("foo")
	top.subscribe(sub)
	if _, ok := top.subs[sub]; !ok {
		t.Errorf("Expect subscription registered with topic on subscribe")
	}
	top.unsubscribe(sub)
	if _, ok := top.subs[sub]; ok {
		t.Errorf("Expect subscription de-registered with topic on unsubscribe")
	}
}

func TestTopicClose(t *testing.T) {
	sub := new(subscriber)
	top := newTopic("foo")
	top.subscribe(sub)
	go func() {
		top.close()
	}()
	select {
	case <-top.done:
	case <-time.After(1 * time.Second):
		t.Errorf("Expect subscription closed")
	}
}
