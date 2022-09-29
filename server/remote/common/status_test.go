package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
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
	proc := &model.Proc{Name: "lint"}

	assert.EqualValues(t, "", GetPipelineStatusContext(repo, pipeline, proc))

	server.Config.Server.StatusContext = "ci/woodpecker"
	server.Config.Server.StatusContextFormat = "{{ .context }}/{{ .event }}/{{ .pipeline }}"
	assert.EqualValues(t, "ci/woodpecker/pr/lint", GetPipelineStatusContext(repo, pipeline, proc))
	pipeline.Event = model.EventPush
	assert.EqualValues(t, "ci/woodpecker/push/lint", GetPipelineStatusContext(repo, pipeline, proc))

	server.Config.Server.StatusContext = "ci"
	server.Config.Server.StatusContextFormat = "{{ .context }}:{{ .owner }}/{{ .repo }}:{{ .event }}:{{ .pipeline }}"
	assert.EqualValues(t, "ci:user1/repo1:push:lint", GetPipelineStatusContext(repo, pipeline, proc))
}
