// Copyright 2024 Woodpecker Authors
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

package atomgit

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"go.woodpecker-ci.org/woodpecker/v3/server/forge/atomgit/fixtures"
	"go.woodpecker-ci.org/woodpecker/v3/server/forge/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

func newHookRequest(event, body string) *http.Request {
	req, _ := http.NewRequest(http.MethodPost, "/hook", bytes.NewBufferString(body))
	req.Header = http.Header{}
	req.Header.Set(hookEvent, event)
	return req
}

func TestParsePushHook(t *testing.T) {
	repo, pipeline, err := parseHook(newHookRequest(hookPush, fixtures.HookPush))
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.NotNil(t, pipeline)
	assert.Equal(t, model.EventPush, pipeline.Event)
	assert.Equal(t, "da1560886d4f094c3e6c9ef40349f7d38b5d27d7", pipeline.Commit)
	assert.Equal(t, "master", pipeline.Branch)
	assert.Equal(t, "test_name/repo_name", repo.FullName)
	assert.Equal(t, []string{"file1.txt", "file2.txt"}, pipeline.ChangedFiles)
}

func TestParseTagPushHook(t *testing.T) {
	repo, pipeline, err := parseHook(newHookRequest(hookTagPush, fixtures.HookTagPush))
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.NotNil(t, pipeline)
	assert.Equal(t, model.EventTag, pipeline.Event)
	assert.Equal(t, "refs/tags/v1.0.0", pipeline.Ref)
	assert.Equal(t, "82b3d5ae55f7080f1e6022629cdb57bfae7cccc7", pipeline.Commit)
}

func TestParseMergeRequestHook(t *testing.T) {
	repo, pipeline, err := parseHook(newHookRequest(hookMergeRequest, fixtures.HookMergeRequest))
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.NotNil(t, pipeline)
	assert.Equal(t, model.EventPull, pipeline.Event)
	assert.Equal(t, "refs/pull/1/head", pipeline.Ref)
	assert.Equal(t, "Add feature", pipeline.Title)
	assert.Equal(t, []string{"bug"}, pipeline.PullRequestLabels)
	assert.Equal(t, "da1560886d4f094c3e6c9ef40349f7d38b5d27d7", pipeline.Commit)
}

func TestParseUnsupportedHook(t *testing.T) {
	req := newHookRequest("unknown_event", "{}")
	_, _, err := parseHook(req)
	assert.Error(t, err)
	var ignore *types.ErrIgnoreEvent
	assert.ErrorAs(t, err, &ignore)
}
