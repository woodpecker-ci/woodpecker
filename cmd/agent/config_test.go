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

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadAgentIDFileNotExists(t *testing.T) {
	assert.EqualValues(t, -1, readAgentConfig("foobar.conf").AgentID)
}

func TestReadAgentIDFileExists(t *testing.T) {
	tmpF, errTmpF := os.CreateTemp("", "tmp_")
	if !assert.NoError(t, errTmpF) {
		t.FailNow()
	}

	errWrite := os.WriteFile(tmpF.Name(), []byte(`{"agent_id":3}`), 0o644)
	if !assert.NoError(t, errWrite) {
		t.FailNow()
	}

	actual := readAgentConfig(tmpF.Name())
	assert.EqualValues(t, AgentConfig{3}, actual)

	actual.AgentID = 33
	writeAgentConfig(actual, tmpF.Name())
	actual = readAgentConfig(tmpF.Name())
	assert.EqualValues(t, 33, actual.AgentID)
}
