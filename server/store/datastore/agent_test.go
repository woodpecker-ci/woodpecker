// Copyright 2023 Woodpecker Authors
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

package datastore

import (
	"testing"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestAgentFindByToken(t *testing.T) {
	store, closer := newTestStore(t, new(model.Agent))
	defer closer()

	agent := &model.Agent{
		ID:    int64(1),
		Name:  "test",
		Token: "secret-token",
	}
	if err := store.AgentCreate(agent); err != nil {
		t.Errorf("Unexpected error: insert agent: %s", err)
		return
	}

	_agent, err := store.AgentFindByToken(agent.Token)
	if err != nil {
		t.Error(err)
		return
	}
	if got, want := _agent.ID, int64(1); got != want {
		t.Errorf("Want config id %d, got %d", want, got)
	}

	_agent, err = store.AgentFindByToken("")
	if err == nil || err.Error() != "Please provide a token" {
		t.Errorf("Expected to get an error for an empty token, but got %s", err)
		return
	}

	if _agent != nil {
		t.Errorf("Expected to not find an agent")
		return
	}
}
