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

package utils

import (
	"sync"
)

// Protected provides thread-safe read and write access to a value of type T.
type Protected[T any] interface {
	// Get returns the current value using a read lock, allowing multiple concurrent
	// readers. Safe to call from multiple goroutines simultaneously.
	Get() T

	// Set replaces the current value using an exclusive write lock.
	// Blocks until all ongoing reads/writes complete.
	Set(v T)

	// Update performs an atomic read-modify-write operation under a single exclusive
	// lock. The provided function receives the current value and returns the new value,
	// eliminating the race condition that would occur with a separate Get + Set.
	Update(fn func(T) T)
}

type protected[T any] struct {
	mu    sync.RWMutex
	value T
}

// NewProtected creates and returns a new Protected wrapper initialized with the
// given value. Use this as the constructor instead of creating a protected struct directly.
func NewProtected[T any](initial T) Protected[T] {
	return &protected[T]{value: initial}
}

func (p *protected[T]) Get() T {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.value
}

func (p *protected[T]) Set(v T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value = v
}

func (p *protected[T]) Update(fn func(T) T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.value = fn(p.value)
}
