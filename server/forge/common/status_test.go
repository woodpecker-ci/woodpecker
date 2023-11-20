// Copyright 2022 Woodpecker Authors
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

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/server"
	"go.woodpecker-ci.org/woodpecker/server/model"
)

func TestGetPipelineStatusContext(t *testing.T) {
	origFormat := server.Config.Server.StatusContextFormat
	origCtx := server.Config.Server.StatusContext
	defer func() {
		server.Config.Server.StatusContextFormat = origFormat
		server.Config.Server.StatusContext = origCtx
	}()

	repo := &model.Repo{Owner: "user1", Name: "repo1"}
	pipeline := &model.Pipeline{Event: model.EventPull}
	workflow := &model.Workflow{Name: "lint"}

	assert.EqualValues(t, "", GetPipelineStatusContext(repo, pipeline, workflow))

	server.Config.Server.StatusContext = "ci/woodpecker"
	server.Config.Server.StatusContextFormat = "{{ .context }}/{{ .event }}/{{ .workflow }}"
	assert.EqualValues(t, "ci/woodpecker/pr/lint", GetPipelineStatusContext(repo, pipeline, workflow))
	pipeline.Event = model.EventPush
	assert.EqualValues(t, "ci/woodpecker/push/lint", GetPipelineStatusContext(repo, pipeline, workflow))

	server.Config.Server.StatusContext = "ci"
	server.Config.Server.StatusContextFormat = "{{ .context }}:{{ .owner }}/{{ .repo }}:{{ .event }}:{{ .workflow }}"
	assert.EqualValues(t, "ci:user1/repo1:push:lint", GetPipelineStatusContext(repo, pipeline, workflow))
}
