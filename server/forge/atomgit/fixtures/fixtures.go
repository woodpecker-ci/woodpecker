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

package fixtures

// HookPush is a AtomGit push webhook payload.
const HookPush = `{
  "object_kind": "push",
  "event_name": "push",
  "before": "95790bf891e76feeb30fa2fcc762bd98c1e28ad9",
  "after": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
  "ref": "refs/heads/master",
  "checkout_sha": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
  "user_id": 4,
  "user_name": "John Doe",
  "user_email": "john@example.com",
  "user_avatar": "https://atomgit.com/avatar.png",
  "project_id": 15,
  "project": {
    "id": 15,
    "name": "repo_name",
    "path_with_namespace": "test_name/repo_name",
    "full_name": "test_name/repo_name",
    "http_url_to_repo": "https://atomgit.com/test_name/repo_name.git",
    "default_branch": "master",
    "html_url": "https://atomgit.com/test_name/repo_name"
  },
  "repository": {
    "id": 15,
    "name": "repo_name",
    "full_name": "test_name/repo_name",
    "http_url_to_repo": "https://atomgit.com/test_name/repo_name.git",
    "default_branch": "master",
    "html_url": "https://atomgit.com/test_name/repo_name"
  },
  "commits": [
    {
      "id": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
      "message": "fix: bugs",
      "title": "fix: bugs",
      "url": "https://atomgit.com/test_name/repo_name/-/commit/da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
      "author": {"name": "John Doe", "email": "john@example.com", "username": "john"},
      "added": ["file1.txt"],
      "modified": ["file2.txt"],
      "removed": []
    }
  ],
  "total_commits_count": 1
}`

// HookTagPush is a AtomGit tag push webhook payload.
const HookTagPush = `{
  "object_kind": "tag_push",
  "event_name": "tag_push",
  "before": "0000000000000000000000000000000000000000",
  "after": "82b3d5ae55f7080f1e6022629cdb57bfae7cccc7",
  "ref": "refs/tags/v1.0.0",
  "checkout_sha": "82b3d5ae55f7080f1e6022629cdb57bfae7cccc7",
  "user_id": 4,
  "user_name": "John Doe",
  "user_email": "john@example.com",
  "user_avatar": "https://atomgit.com/avatar.png",
  "project_id": 15,
  "project": {
    "id": 15,
    "name": "repo_name",
    "full_name": "test_name/repo_name",
    "http_url_to_repo": "https://atomgit.com/test_name/repo_name.git",
    "default_branch": "master",
    "html_url": "https://atomgit.com/test_name/repo_name"
  },
  "repository": {
    "id": 15,
    "name": "repo_name",
    "full_name": "test_name/repo_name",
    "http_url_to_repo": "https://atomgit.com/test_name/repo_name.git",
    "default_branch": "master",
    "html_url": "https://atomgit.com/test_name/repo_name"
  },
  "commits": [],
  "total_commits_count": 0
}`

// HookMergeRequest is a AtomGit merge request (open) webhook payload.
const HookMergeRequest = `{
  "object_kind": "merge_request",
  "event_name": "merge_request_open",
  "user": {
    "id": 1,
    "username": "someuser",
    "name": "Some User",
    "email": "someuser@atomgit.com",
    "avatar_url": "https://atomgit.com/avatar.png",
    "html_url": "https://atomgit.com/someuser"
  },
  "project": {
    "id": 15,
    "name": "repo_name",
    "full_name": "test_name/repo_name",
    "http_url_to_repo": "https://atomgit.com/test_name/repo_name.git",
    "default_branch": "master",
    "html_url": "https://atomgit.com/test_name/repo_name"
  },
  "repository": {
    "id": 15,
    "name": "repo_name",
    "full_name": "test_name/repo_name",
    "http_url_to_repo": "https://atomgit.com/test_name/repo_name.git",
    "default_branch": "master",
    "html_url": "https://atomgit.com/test_name/repo_name"
  },
  "object_attributes": {
    "id": 99,
    "iid": 1,
    "target_branch": "master",
    "source_branch": "feature",
    "title": "Add feature",
    "state": "opened",
    "html_url": "https://atomgit.com/test_name/repo_name/-/merge_requests/1",
    "author": {
      "id": 1,
      "username": "someuser",
      "name": "Some User",
      "email": "someuser@atomgit.com",
      "avatar_url": "https://atomgit.com/avatar.png"
    },
    "source_repo": {
      "id": 15,
      "full_name": "test_name/repo_name"
    },
    "target_repo": {
      "id": 15,
      "full_name": "test_name/repo_name"
    },
    "draft": false,
    "last_commit": {
      "id": "da1560886d4f094c3e6c9ef40349f7d38b5d27d7"
    }
  },
  "labels": [
    {"id": 1, "name": "bug", "color": "#d9534f"}
  ]
}`
