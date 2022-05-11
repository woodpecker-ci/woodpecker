package common

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

func TestGetBuildStatusContext(t *testing.T) {
	origFormatr := server.Config.Server.StatusContextFormat
	origCtx := server.Config.Server.StatusContext
	defer func() {
		server.Config.Server.StatusContextFormat = origFormatr
		server.Config.Server.StatusContext = origCtx
	}()

	repo := &model.Repo{Owner: "user1", Name: "repo1"}
	build := &model.Build{Event: model.EventPull}
	proc := &model.Proc{Name: "lint"}

	assert.EqualValues(t, "", GetBuildStatusContext(repo, build, proc))

	server.Config.Server.StatusContext = "ci/woodpecker"
	server.Config.Server.StatusContextFormat = "%context/%event/%pipeline"
	assert.EqualValues(t, "ci/woodpecker/pr/lint", GetBuildStatusContext(repo, build, proc))
	build.Event = model.EventPush
	assert.EqualValues(t, "ci/woodpecker/push/lint", GetBuildStatusContext(repo, build, proc))

	server.Config.Server.StatusContext = "ci"
	server.Config.Server.StatusContextFormat = "%context:%owner/%repo:%event:%pipeline"
	assert.EqualValues(t, "ci:user1/repo1:push:lint", GetBuildStatusContext(repo, build, proc))
}
