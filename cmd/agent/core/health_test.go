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

package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v2/agent"
)

func TestHealthy(t *testing.T) {
	s := agent.State{}
	s.Metadata = map[string]agent.Info{}

	s.Add("1", time.Hour, "octocat/hello-world", "42")

	assert.Equal(t, "1", s.Metadata["1"].ID)
	assert.Equal(t, time.Hour, s.Metadata["1"].Timeout)
	assert.Equal(t, "octocat/hello-world", s.Metadata["1"].Repo)

	s.Metadata["1"] = agent.Info{
		Timeout: time.Hour,
		Started: time.Now().UTC(),
	}
	assert.True(t, s.Healthy(), "want healthy status when timeout not exceeded, got false")

	s.Metadata["1"] = agent.Info{
		Started: time.Now().UTC().Add(-(time.Minute * 30)),
	}
	assert.True(t, s.Healthy(), "want healthy status when timeout+buffer not exceeded, got false")

	s.Metadata["1"] = agent.Info{
		Started: time.Now().UTC().Add(-(time.Hour + time.Minute)),
	}
	assert.False(t, s.Healthy(), "want unhealthy status when timeout+buffer not exceeded, got true")
}
