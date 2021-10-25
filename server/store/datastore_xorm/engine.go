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

package datastore_xorm

import (
	"github.com/woodpecker-ci/woodpecker/server/store"

	"github.com/rs/zerolog/log"
	"xorm.io/xorm"
)

type storage struct {
	engine *xorm.Engine
}

// make sure storage implement Store
var _ store.Store = &storage{}

func init() {
	store.RegisterAdapter(newEngine, "xorm")
}

func newEngine(opts *store.Opts) (store.Store, error) {
	engine, err := xorm.NewEngine(opts.Driver, opts.Config)
	if err != nil {
		return nil, err
	}
	engine.SetLogger(log.Logger) // TODO: special config to enable/disable ?
	return &storage{
		engine: engine,
	}, nil
}
