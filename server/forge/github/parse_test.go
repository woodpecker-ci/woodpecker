// Copyright 2022 Woodpecker Authors
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

package github

import (
	"bytes"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/github/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

const (
	hookEvent   = "X-GitHub-Event"
	hookDeploy  = "deployment"
	hookPush    = "push"
	hookPull    = "pull_request"
	hookRelease = "release"
)

func testHookRequest(payload []byte, event string) *http.Request {
	buf := bytes.NewBuffer(payload)
	req, _ := http.NewRequest(http.MethodPost, "/hook", buf)
	req.Header = http.Header{}
	req.Header.Set(hookEvent, event)
	return req
}

func Test_parseHook(t *testing.T) {
	t.Run("ignore unsupported hook events", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequest), "issues")
		p, r, b, err := parseHook(req, false)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.Nil(t, p)
		assert.ErrorIs(t, err, &types.ErrIgnoreEvent{})
	})

	t.Run("skip skip push hook when action is deleted", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPushDeleted), hookPush)
		p, r, b, err := parseHook(req, false)
		assert.Nil(t, r)
		assert.Nil(t, b)
		assert.NoError(t, err)
		assert.Nil(t, p)
	})
	t.Run("push hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPush), hookPush)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.Nil(t, p)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Equal(t, model.EventPush, b.Event)
		sort.Strings(b.ChangedFiles)
		assert.Equal(t, []string{"pipeline/shared/replace_secrets.go", "pipeline/shared/replace_secrets_test.go"}, b.ChangedFiles)
	})

	t.Run("PR hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequest), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.NotNil(t, p)
		assert.Equal(t, model.EventPull, b.Event)
	})
	t.Run("PR closed hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestClosed), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.NotNil(t, p)
		assert.Equal(t, model.EventPullClosed, b.Event)
	})
	t.Run("PR merged hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookPullRequestMerged), hookPull)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.NotNil(t, p)
		assert.Equal(t, model.EventPullClosed, b.Event)
	})

	t.Run("deploy hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookDeploy), hookDeploy)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Nil(t, p)
		assert.Equal(t, model.EventDeploy, b.Event)
		assert.Equal(t, "production", b.DeployTo)
		assert.Equal(t, "deploy", b.DeployTask)
	})

	t.Run("release hook", func(t *testing.T) {
		req := testHookRequest([]byte(fixtures.HookRelease), hookRelease)
		p, r, b, err := parseHook(req, false)
		assert.NoError(t, err)
		assert.NotNil(t, r)
		assert.NotNil(t, b)
		assert.Nil(t, p)
		assert.Equal(t, model.EventRelease, b.Event)
		assert.Len(t, strings.Split(b.Ref, "/"), 3)
		assert.True(t, strings.HasPrefix(b.Ref, "refs/tags/"))
	})
}
