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

package wasm

import (
	"context"
	"sync"
)

// keyedLock is a set of locks that are created when needed and forgotten once nobody uses them anymore.
type keyedLock struct {
	mu    sync.Mutex
	locks map[string]*keyLock
}

type keyLock struct {
	held  chan struct{}
	users int
}

// lock waits until the lock of the key is free or the context is done.
func (k *keyedLock) lock(ctx context.Context, key string) (unlock func(), err error) {
	k.mu.Lock()
	if k.locks == nil {
		k.locks = make(map[string]*keyLock)
	}
	l, ok := k.locks[key]
	if !ok {
		l = &keyLock{held: make(chan struct{}, 1)}
		k.locks[key] = l
	}
	l.users++
	k.mu.Unlock()

	forget := func() {
		k.mu.Lock()
		defer k.mu.Unlock()
		if l.users--; l.users == 0 {
			delete(k.locks, key)
		}
	}

	select {
	case l.held <- struct{}{}:
		return func() {
			<-l.held
			forget()
		}, nil
	case <-ctx.Done():
		forget()
		return nil, ctx.Err()
	}
}
