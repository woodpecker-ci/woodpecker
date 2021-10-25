// Copyright 2021 Woodpecker Authors
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
	"fmt"

	"github.com/rs/zerolog/log"
)

// Opts are options for a new database connection
type Opts struct {
	Driver  string
	Config  string
	Adapter string
}

var storeCreators map[string]func(opts *Opts) (Store, error)

// New creates a database connection for the given driver and datasource
// and returns a new Store.
func New(opts *Opts) (Store, error) {
	if fn, ok := storeCreators[opts.Adapter]; ok {
		return fn(opts)
	}
	return nil, fmt.Errorf("adapter '%s' not found", opts.Adapter)
}

func init() {
	storeCreators = make(map[string]func(opts *Opts) (Store, error))
}

func RegisterAdapter(fn func(opts *Opts) (Store, error), name string) {
	log.Info().Msgf("Add adapter '%s'", name)
	storeCreators[name] = fn
}
