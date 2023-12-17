// Copyright 2018 Drone.IO Inc.
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

package store

import (
	"context"
)

const key = "store"

// Setter defines a context that enables setting values.
type Setter interface {
	Set(string, any)
}

// FromContext returns the Store associated with this context.
func FromContext(c context.Context) Store {
	store, _ := c.Value(key).(Store)
	return store
}

// TryFromContext try to return the Store associated with this context.
func TryFromContext(c context.Context) (Store, bool) {
	store, ok := c.Value(key).(Store)
	return store, ok
}

// ToContext adds the Store to this context if it supports
// the Setter interface.
func ToContext(c Setter, store Store) {
	c.Set(key, store)
}

func InjectToContext(ctx context.Context, store Store) context.Context {
	return context.WithValue(ctx, key, store)
}
