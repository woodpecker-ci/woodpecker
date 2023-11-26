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

package datastore

import (
	"testing"

	"go.woodpecker-ci.org/woodpecker/server/model"
)

func TestServerConfigGetSet(t *testing.T) {
	store, closer := newTestStore(t, new(model.ServerConfig))
	defer closer()

	serverConfig := &model.ServerConfig{
		Key:   "test",
		Value: "wonderland",
	}
	if err := store.ServerConfigSet(serverConfig.Key, serverConfig.Value); err != nil {
		t.Errorf("Unexpected error: insert secret: %s", err)
		return
	}

	value, err := store.ServerConfigGet(serverConfig.Key)
	if err != nil {
		t.Errorf("Unexpected error: delete secret: %s", err)
		return
	}

	if value != serverConfig.Value {
		t.Errorf("Want server-config value %s, got %s", serverConfig.Value, value)
		return
	}

	serverConfig.Value = "new-wonderland"
	if err := store.ServerConfigSet(serverConfig.Key, serverConfig.Value); err != nil {
		t.Errorf("Unexpected error: insert secret: %s", err)
		return
	}

	value, err = store.ServerConfigGet(serverConfig.Key)
	if err != nil {
		t.Errorf("Unexpected error: delete secret: %s", err)
		return
	}

	if value != serverConfig.Value {
		t.Errorf("Want server-config value %s, got %s", serverConfig.Value, value)
		return
	}

	value, err = store.ServerConfigGet("config_not_exist")
	if err == nil {
		t.Errorf("Unexpected: no error on missing config: %v", err)
		return
	}
	if value != "" {
		t.Errorf("Unexpected: got value on missing config: %s", value)
		return
	}
}
