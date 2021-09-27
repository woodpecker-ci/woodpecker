// Copyright 2021 Woodpecker Authors
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

package testdata

import (
	"net/http"
	"net/url"
)

var (
	ServiceHookMethod = http.MethodPost
	ServiceHookURL, _ = url.Parse(
		"http://10.40.8.5:8000/hook?owner=test&name=woodpecker&access_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9." +
			"eyJ0ZXh0IjoidGVzdC93b29kcGVja2VyIiwidHlwZSI6Imhvb2sifQ.x3kPnmZtxZQ_9_eMhfQ1HSmj_SLhdT_Lu2hMczWjKh0")
	ServiceHookHeaders = http.Header{
		"Content-Type":   []string{"application/json"},
		"User-Agent":     []string{"GitLab/14.3.0"},
		"X-Gitlab-Event": []string{"Service Hook"},
	}
)

// ServiceHookPushBody is payload of ServiceHook: Push
var ServiceHookPushBody = `{
  "object_kind": "push",
  "event_name": "push",
  "before": "ffe8eb4f91d1fe6bc49f1e610e50e4b5767f0104",
  "after": "16862e368d8ab812e48833b741dad720d6e2cb7f",
  "ref": "refs/heads/master",
  "checkout_sha": "16862e368d8ab812e48833b741dad720d6e2cb7f",
  "message": null,
  "user_id": 2,
  "user_name": "te st",
  "user_username": "test",
  "user_email": "",
  "user_avatar": "https://www.gravatar.com/avatar/dd46a756faad4727fb679320751f6dea?s=80&d=identicon",
  "project_id": 2,
  "project": {
    "id": 2,
    "name": "Woodpecker",
    "description": "",
    "web_url": "http://10.40.8.5:3200/test/woodpecker",
    "avatar_url": null,
    "git_ssh_url": "git@10.40.8.5:test/woodpecker.git",
    "git_http_url": "http://10.40.8.5:3200/test/woodpecker.git",
    "namespace": "te st",
    "visibility_level": 20,
    "path_with_namespace": "test/woodpecker",
    "default_branch": "master",
    "ci_config_path": null,
    "homepage": "http://10.40.8.5:3200/test/woodpecker",
    "url": "git@10.40.8.5:test/woodpecker.git",
    "ssh_url": "git@10.40.8.5:test/woodpecker.git",
    "http_url": "http://10.40.8.5:3200/test/woodpecker.git"
  },
  "commits": [
    {
      "id": "16862e368d8ab812e48833b741dad720d6e2cb7f",
      "message": "Update main.go",
      "title": "Update main.go",
      "timestamp": "2021-09-27T04:46:14+00:00",
      "url": "http://10.40.8.5:3200/test/woodpecker/-/commit/16862e368d8ab812e48833b741dad720d6e2cb7f",
      "author": {
        "name": "te st",
        "email": "test@test.test"
      },
      "added": [

      ],
      "modified": [
        "cmd/cli/main.go"
      ],
      "removed": [

      ]
    }
  ],
  "total_commits_count": 1,
  "push_options": {
  },
  "repository": {
    "name": "Woodpecker",
    "url": "git@10.40.8.5:test/woodpecker.git",
    "description": "",
    "homepage": "http://10.40.8.5:3200/test/woodpecker",
    "git_http_url": "http://10.40.8.5:3200/test/woodpecker.git",
    "git_ssh_url": "git@10.40.8.5:test/woodpecker.git",
    "visibility_level": 20
  }
}`

// ServiceHookTagPushBody is payload of ServiceHook: TagPush
var ServiceHookTagPushBody = `{
  "object_kind": "tag_push",
  "event_name": "tag_push",
  "before": "0000000000000000000000000000000000000000",
  "after": "fabed3d94cd03e6c2b7958afa9569c18a24d301f",
  "ref": "refs/tags/v22",
  "checkout_sha": "16862e368d8ab812e48833b741dad720d6e2cb7f",
  "message": "hi",
  "user_id": 2,
  "user_name": "te st",
  "user_username": "test",
  "user_email": "",
  "user_avatar": "https://www.gravatar.com/avatar/dd46a756faad4727fb679320751f6dea?s=80&d=identicon",
  "project_id": 2,
  "project": {
    "id": 2,
    "name": "Woodpecker",
    "description": "",
    "web_url": "http://10.40.8.5:3200/test/woodpecker",
    "avatar_url": null,
    "git_ssh_url": "git@10.40.8.5:test/woodpecker.git",
    "git_http_url": "http://10.40.8.5:3200/test/woodpecker.git",
    "namespace": "te st",
    "visibility_level": 20,
    "path_with_namespace": "test/woodpecker",
    "default_branch": "master",
    "ci_config_path": null,
    "homepage": "http://10.40.8.5:3200/test/woodpecker",
    "url": "git@10.40.8.5:test/woodpecker.git",
    "ssh_url": "git@10.40.8.5:test/woodpecker.git",
    "http_url": "http://10.40.8.5:3200/test/woodpecker.git"
  },
  "commits": [
    {
      "id": "16862e368d8ab812e48833b741dad720d6e2cb7f",
      "message": "Update main.go",
      "title": "Update main.go",
      "timestamp": "2021-09-27T04:46:14+00:00",
      "url": "http://10.40.8.5:3200/test/woodpecker/-/commit/16862e368d8ab812e48833b741dad720d6e2cb7f",
      "author": {
        "name": "te st",
        "email": "test@test.test"
      },
      "added": [

      ],
      "modified": [
        "cmd/cli/main.go"
      ],
      "removed": [

      ]
    }
  ],
  "total_commits_count": 1,
  "push_options": {
  },
  "repository": {
    "name": "Woodpecker",
    "url": "git@10.40.8.5:test/woodpecker.git",
    "description": "",
    "homepage": "http://10.40.8.5:3200/test/woodpecker",
    "git_http_url": "http://10.40.8.5:3200/test/woodpecker.git",
    "git_ssh_url": "git@10.40.8.5:test/woodpecker.git",
    "visibility_level": 20
  }
}`

// ServiceHookMergeRequestBody is payload of ServiceHook: MergeRequest
var ServiceHookMergeRequestBody = `{
  "object_kind": "tag_push",
  "event_name": "tag_push",
  "before": "0000000000000000000000000000000000000000",
  "after": "fabed3d94cd03e6c2b7958afa9569c18a24d301f",
  "ref": "refs/tags/v22",
  "checkout_sha": "16862e368d8ab812e48833b741dad720d6e2cb7f",
  "message": "hi",
  "user_id": 2,
  "user_name": "te st",
  "user_username": "test",
  "user_email": "",
  "user_avatar": "https://www.gravatar.com/avatar/dd46a756faad4727fb679320751f6dea?s=80&d=identicon",
  "project_id": 2,
  "project": {
    "id": 2,
    "name": "Woodpecker",
    "description": "",
    "web_url": "http://10.40.8.5:3200/test/woodpecker",
    "avatar_url": null,
    "git_ssh_url": "git@10.40.8.5:test/woodpecker.git",
    "git_http_url": "http://10.40.8.5:3200/test/woodpecker.git",
    "namespace": "te st",
    "visibility_level": 20,
    "path_with_namespace": "test/woodpecker",
    "default_branch": "master",
    "ci_config_path": null,
    "homepage": "http://10.40.8.5:3200/test/woodpecker",
    "url": "git@10.40.8.5:test/woodpecker.git",
    "ssh_url": "git@10.40.8.5:test/woodpecker.git",
    "http_url": "http://10.40.8.5:3200/test/woodpecker.git"
  },
  "commits": [
    {
      "id": "16862e368d8ab812e48833b741dad720d6e2cb7f",
      "message": "Update main.go",
      "title": "Update main.go",
      "timestamp": "2021-09-27T04:46:14+00:00",
      "url": "http://10.40.8.5:3200/test/woodpecker/-/commit/16862e368d8ab812e48833b741dad720d6e2cb7f",
      "author": {
        "name": "te st",
        "email": "test@test.test"
      },
      "added": [

      ],
      "modified": [
        "cmd/cli/main.go"
      ],
      "removed": [

      ]
    }
  ],
  "total_commits_count": 1,
  "push_options": {
  },
  "repository": {
    "name": "Woodpecker",
    "url": "git@10.40.8.5:test/woodpecker.git",
    "description": "",
    "homepage": "http://10.40.8.5:3200/test/woodpecker",
    "git_http_url": "http://10.40.8.5:3200/test/woodpecker.git",
    "git_ssh_url": "git@10.40.8.5:test/woodpecker.git",
    "visibility_level": 20
  }
}`
