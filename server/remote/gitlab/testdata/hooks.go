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
var ServiceHookPushBody = []byte(`{
  "object_kind": "push",
  "event_name": "push",
  "before": "ffe8eb4f91d1fe6bc49f1e610e50e4b5767f0104",
  "after": "16862e368d8ab812e48833b741dad720d6e2cb7f",
  "ref": "refs/heads/master",
  "checkout_sha": "16862e368d8ab812e48833b741dad720d6e2cb7f",
  "message": null,
  "user_id": 2,
  "user_name": "the test",
  "user_username": "test",
  "user_email": "",
  "user_avatar": "https://www.gravatar.com/avatar/dd46a756faad4727fb679320751f6dea?s=80&d=identicon",
  "project_id": 2,
  "project": {
    "id": 2,
    "name": "Woodpecker",
    "description": "",
    "web_url": "http://10.40.8.5:3200/test/woodpecker",
    "avatar_url": "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg",
    "git_ssh_url": "git@10.40.8.5:test/woodpecker.git",
    "git_http_url": "http://10.40.8.5:3200/test/woodpecker.git",
    "namespace": "the test",
    "visibility_level": 20,
    "path_with_namespace": "test/woodpecker",
    "default_branch": "develop",
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
        "name": "the test",
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
}`)

// ServiceHookTagPushBody is payload of ServiceHook: TagPush
var ServiceHookTagPushBody = []byte(`{
  "object_kind": "tag_push",
  "event_name": "tag_push",
  "before": "0000000000000000000000000000000000000000",
  "after": "fabed3d94cd03e6c2b7958afa9569c18a24d301f",
  "ref": "refs/tags/v22",
  "checkout_sha": "16862e368d8ab812e48833b741dad720d6e2cb7f",
  "message": "hi",
  "user_id": 2,
  "user_name": "the test",
  "user_username": "test",
  "user_email": "",
  "user_avatar": "https://www.gravatar.com/avatar/dd46a756faad4727fb679320751f6dea?s=80&d=identicon",
  "project_id": 2,
  "project": {
    "id": 2,
    "name": "Woodpecker",
    "description": "",
    "web_url": "http://10.40.8.5:3200/test/woodpecker",
    "avatar_url": "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg",
    "git_ssh_url": "git@10.40.8.5:test/woodpecker.git",
    "git_http_url": "http://10.40.8.5:3200/test/woodpecker.git",
    "namespace": "the test",
    "visibility_level": 20,
    "path_with_namespace": "test/woodpecker",
    "default_branch": "develop",
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
        "name": "the test",
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
}`)

// WebhookMergeRequestBody is payload of MergeEvent
var WebhookMergeRequestBody = []byte(`{
  "object_kind": "merge_request",
  "event_type": "merge_request",
  "user": {
    "id": 2251488,
    "name": "Anbraten",
    "username": "anbraten",
    "avatar_url": "https://secure.gravatar.com/avatar/fc9b6fe77c6b732a02925a62a81f05a0?s=80&d=identicon",
    "email": "some@mail.info"
  },
  "project": {
    "id": 32059612,
    "name": "woodpecker",
    "description": "",
    "web_url": "https://gitlab.com/anbraten/woodpecker",
    "avatar_url": "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg",
    "git_ssh_url": "git@gitlab.com:anbraten/woodpecker.git",
    "git_http_url": "https://gitlab.com/anbraten/woodpecker.git",
    "namespace": "Anbraten",
    "visibility_level": 20,
    "path_with_namespace": "anbraten/woodpecker",
    "default_branch": "main",
    "ci_config_path": "",
    "homepage": "https://gitlab.com/anbraten/woodpecker",
    "url": "git@gitlab.com:anbraten/woodpecker.git",
    "ssh_url": "git@gitlab.com:anbraten/woodpecker.git",
    "http_url": "https://gitlab.com/anbraten/woodpecker.git"
  },
  "object_attributes": {
    "assignee_id": 2251488,
    "author_id": 2251488,
    "created_at": "2022-01-10 15:23:41 UTC",
    "description": "",
    "head_pipeline_id": 449733536,
    "id": 134400602,
    "iid": 3,
    "last_edited_at": "2022-01-17 15:46:23 UTC",
    "last_edited_by_id": 2251488,
    "merge_commit_sha": null,
    "merge_error": null,
    "merge_params": {
      "force_remove_source_branch": "1"
    },
    "merge_status": "unchecked",
    "merge_user_id": null,
    "merge_when_pipeline_succeeds": false,
    "milestone_id": null,
    "source_branch": "anbraten-main-patch-05373",
    "source_project_id": 32059612,
    "state_id": 1,
    "target_branch": "main",
    "target_project_id": 32059612,
    "time_estimate": 0,
    "title": "Update client.go 🎉",
    "updated_at": "2022-01-17 15:47:39 UTC",
    "updated_by_id": 2251488,
    "url": "https://gitlab.com/anbraten/woodpecker/-/merge_requests/3",
    "source": {
      "id": 32059612,
      "name": "woodpecker",
      "description": "",
      "web_url": "https://gitlab.com/anbraten/woodpecker",
      "avatar_url": null,
      "git_ssh_url": "git@gitlab.com:anbraten/woodpecker.git",
      "git_http_url": "https://gitlab.com/anbraten/woodpecker.git",
      "namespace": "Anbraten",
      "visibility_level": 20,
      "path_with_namespace": "anbraten/woodpecker",
      "default_branch": "main",
      "ci_config_path": "",
      "homepage": "https://gitlab.com/anbraten/woodpecker",
      "url": "git@gitlab.com:anbraten/woodpecker.git",
      "ssh_url": "git@gitlab.com:anbraten/woodpecker.git",
      "http_url": "https://gitlab.com/anbraten/woodpecker.git"
    },
    "target": {
      "id": 32059612,
      "name": "woodpecker",
      "description": "",
      "web_url": "https://gitlab.com/anbraten/woodpecker",
      "avatar_url": "http://example.com/uploads/project/avatar/555/Outh-20-Logo.jpg",
      "git_ssh_url": "git@gitlab.com:anbraten/woodpecker.git",
      "git_http_url": "https://gitlab.com/anbraten/woodpecker.git",
      "namespace": "Anbraten",
      "visibility_level": 20,
      "path_with_namespace": "anbraten/woodpecker",
      "default_branch": "main",
      "ci_config_path": "",
      "homepage": "https://gitlab.com/anbraten/woodpecker",
      "url": "git@gitlab.com:anbraten/woodpecker.git",
      "ssh_url": "git@gitlab.com:anbraten/woodpecker.git",
      "http_url": "https://gitlab.com/anbraten/woodpecker.git"
    },
    "last_commit": {
      "id": "c136499ec574e1034b24c5d306de9acda3005367",
      "message": "Update folder/todo.txt",
      "title": "Update folder/todo.txt",
      "timestamp": "2022-01-17T15:47:38+00:00",
      "url": "https://gitlab.com/anbraten/woodpecker/-/commit/c136499ec574e1034b24c5d306de9acda3005367",
      "author": {
        "name": "Anbraten",
        "email": "some@mail.info"
      }
    },
    "work_in_progress": false,
    "total_time_spent": 0,
    "time_change": 0,
    "human_total_time_spent": null,
    "human_time_change": null,
    "human_time_estimate": null,
    "assignee_ids": [
      2251488
    ],
    "state": "opened",
    "blocking_discussions_resolved": true,
    "action": "update",
    "oldrev": "8b641937b7340066d882b9d8a8cc5b0573a207de"
  },
  "labels": [

  ],
  "changes": {
    "updated_at": {
      "previous": "2022-01-17 15:46:23 UTC",
      "current": "2022-01-17 15:47:39 UTC"
    }
  },
  "repository": {
    "name": "woodpecker",
    "url": "git@gitlab.com:anbraten/woodpecker.git",
    "description": "",
    "homepage": "https://gitlab.com/anbraten/woodpecker"
  },
  "assignees": [
    {
      "id": 2251488,
      "name": "Anbraten",
      "username": "anbraten",
      "avatar_url": "https://secure.gravatar.com/avatar/fc9b6fe77c6b732a02925a62a81f05a0?s=80&d=identicon",
      "email": "some@mail.info"
    }
  ]
}
`)

var WebhookReleqaseBody = []byte(`
{
  "id": 4268085,
  "created_at": "2022-02-09 20:19:09 UTC",
  "description": "new version desc",
  "name": "Awesome version 0.0.2",
  "released_at": "2022-02-09 20:19:09 UTC",
  "tag": "0.0.2",
  "object_kind": "release",
  "project": {
    "id": 32521798,
    "name": "ci",
    "description": "",
    "web_url": "https://gitlab.com/anbratens-test/ci",
    "avatar_url": null,
    "git_ssh_url": "git@gitlab.com:anbratens-test/ci.git",
    "git_http_url": "https://gitlab.com/anbratens-test/ci.git",
    "namespace": "anbratens-test",
    "visibility_level": 0,
    "path_with_namespace": "anbratens-test/ci",
    "default_branch": "main",
    "ci_config_path": "",
    "homepage": "https://gitlab.com/anbratens-test/ci",
    "url": "git@gitlab.com:anbratens-test/ci.git",
    "ssh_url": "git@gitlab.com:anbratens-test/ci.git",
    "http_url": "https://gitlab.com/anbratens-test/ci.git"
  },
  "url": "https://gitlab.com/anbratens-test/ci/-/releases/0.0.2",
  "action": "create",
  "assets": {
    "count": 4,
    "links": [

    ],
    "sources": [
      {
        "format": "zip",
        "url": "https://gitlab.com/anbratens-test/ci/-/archive/0.0.2/ci-0.0.2.zip"
      },
      {
        "format": "tar.gz",
        "url": "https://gitlab.com/anbratens-test/ci/-/archive/0.0.2/ci-0.0.2.tar.gz"
      },
      {
        "format": "tar.bz2",
        "url": "https://gitlab.com/anbratens-test/ci/-/archive/0.0.2/ci-0.0.2.tar.bz2"
      },
      {
        "format": "tar",
        "url": "https://gitlab.com/anbratens-test/ci/-/archive/0.0.2/ci-0.0.2.tar"
      }
    ]
  },
  "commit": {
    "id": "0b8c02955ba445ea70d22824d9589678852e2b93",
    "message": "Initial commit",
    "title": "Initial commit",
    "timestamp": "2022-01-03T10:39:51+00:00",
    "url": "https://gitlab.com/anbratens-test/ci/-/commit/0b8c02955ba445ea70d22824d9589678852e2b93",
    "author": {
      "name": "Anbraten",
      "email": "2251488-anbraten@users.noreply.gitlab.com"
    }
  }
}
`)
