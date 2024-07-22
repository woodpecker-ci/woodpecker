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

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

func TestServerConfigGetSet(t *testing.T) {
	store, closer := newTestStore(t, new(model.ServerConfig))
	defer closer()

	serverConfig := &model.ServerConfig{
		Key:   "test",
		Value: "wonderland",
	}
	assert.NoError(t, store.ServerConfigSet(serverConfig.Key, serverConfig.Value))

	value, err := store.ServerConfigGet(serverConfig.Key)
	assert.NoError(t, err)
	assert.Equal(t, serverConfig.Value, value)

	serverConfig.Value = "new-wonderland"
	assert.NoError(t, store.ServerConfigSet(serverConfig.Key, serverConfig.Value))

	value, err = store.ServerConfigGet(serverConfig.Key)
	assert.NoError(t, err)
	assert.Equal(t, serverConfig.Value, value)

	value, err = store.ServerConfigGet("config_not_exist")
	assert.Error(t, err)
	assert.Empty(t, value)
}
