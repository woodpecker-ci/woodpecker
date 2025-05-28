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

package fixtures

import _ "embed"

// HookPush is a sample Gitea push hook.
//
//go:embed HookPush.json
var HookPush string

// HookPushMulti push multible commits to a branch.
//
//go:embed HookPushMulti.json
var HookPushMulti string

// HookPushBranch is a sample Gitea push hook where a new branch was created from an existing commit.
//
//go:embed HookPushBranch.json
var HookPushBranch string

// HookTag is a sample Gitea tag hook.
//
//go:embed HookTag.json
var HookTag string

// HookPullRequest is a sample pull_request webhook payload.
//
//go:embed HookPullRequest.json
var HookPullRequest string

//go:embed HookPullRequestUpdated.json
var HookPullRequestUpdated string

//go:embed HookPullRequestMerged.json
var HookPullRequestMerged string

//go:embed HookPullRequestClosed.json
var HookPullRequestClosed string

const HookPullRequestChangeTitleHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request
`

const HookPullRequestChangeTitle = `{
  "action": "edited",
  "number": 7,
  "changes": {
    "title": {
      "from": "Update .woodpecker.yml"
    }
  },
  "pull_request": {
    "id": 3779,
    "url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "number": 7,
    "user": {
      "id": 21,
      "login": "jony",
      "full_name": "Jony",
      "email": "jony@noreply.example.org",
      "avatar_url": "https://gitea.com/avatars/81027235e996f5e3ef6257152357b85d94171a2e",
      "html_url": "https://gitea.com/jony",
      "last_login": "0001-01-01T00:00:00Z",
      "created": "2018-01-25T14:38:19+01:00",
      "visibility": "public",
      "username": "jony"
    },
    "title": "Edit pull title :D",
    "body": "",
    "labels": [],
    "milestone": null,
    "assignees": null,
    "requested_reviewers": null,
    "state": "open",
    "additions": 1,
    "deletions": 0,
    "changed_files": 1,
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "diff_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.diff",
    "patch_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.patch",
    "base": {
      "label": "main",
      "ref": "main",
      "sha": "a40211c506550ebd79633d84e913dafa184c6d56",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "head": {
      "label": "jony-patch-1",
      "ref": "jony-patch-1",
      "sha": "07977177c2cd7d46bad37b8472a9d50e7acb9d1f",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "merge_base": "a40211c506550ebd79633d84e913dafa184c6d56",
    "due_date": null,
    "closed_at": null,
    "pin_order": 0
  },
  "requested_reviewer": null,
  "repository": {
    "id": 1234,
    "owner": {
      "id": 8765,
      "login": "a_nice_user",
      "full_name": "Nice User",
      "email": "a_nice_user@me.mail",
      "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
      "html_url": "https://gitea.com/a_nice_user",
      "created": "2023-05-23T15:17:35+02:00",
      "visibility": "public",
      "username": "a_nice_user"
    },
    "name": "hello_world_ci",
    "full_name": "a_nice_user/hello_world_ci",
    "private": false,
    "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
    "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
    "link": "",
    "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
    "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
    "default_branch": "main",
    "permissions": {
      "admin": true,
      "push": true,
      "pull": true
    },
    "object_format_name": "sha1",
  },
  "sender": {
    "id": 8765,
    "login": "a_nice_user",
    "full_name": "Nice User",
    "email": "a_nice_user@noreply.example.org",
    "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
    "html_url": "https://gitea.com/a_nice_user",
    "created": "2023-05-23T15:17:35+02:00",
    "visibility": "public",
    "username": "a_nice_user"
  },
  "review": null
}`

const HookPullRequestChangeBodyHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request
`

const HookPullRequestChangeBody = `{
  "action": "edited",
  "number": 7,
  "changes": {
    "body": {
      "from": ""
    }
  },
  "pull_request": {
    "id": 3779,
    "url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "number": 7,
    "user": {
      "id": 21,
      "login": "jony",
      "full_name": "Jony",
      "email": "jony@noreply.example.org",
      "avatar_url": "https://gitea.com/avatars/81027235e996f5e3ef6257152357b85d94171a2e",
      "html_url": "https://gitea.com/jony",
      "created": "2018-01-25T14:38:19+01:00",
      "visibility": "public",
      "username": "jony"
    },
    "title": "somepull",
    "body": "wow aaa new pulll body",
    "labels": [],
    "milestone": null,
    "assignees": null,
    "requested_reviewers": null,
    "state": "open",
    "additions": 1,
    "deletions": 0,
    "changed_files": 1,
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "diff_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.diff",
    "patch_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.patch",
    "base": {
      "label": "main",
      "ref": "main",
      "sha": "a40211c506550ebd79633d84e913dafa184c6d56",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "head": {
      "label": "jony-patch-1",
      "ref": "jony-patch-1",
      "sha": "07977177c2cd7d46bad37b8472a9d50e7acb9d1f",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "merge_base": "a40211c506550ebd79633d84e913dafa184c6d56",
    "due_date": null,
    "closed_at": null,
    "pin_order": 0
  },
  "requested_reviewer": null,
  "repository": {
    "id": 1234,
    "owner": {
      "id": 8765,
      "login": "a_nice_user",
      "full_name": "Nice User",
      "email": "a_nice_user@me.mail",
      "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
      "html_url": "https://gitea.com/a_nice_user",
      "created": "2023-05-23T15:17:35+02:00",
      "visibility": "public",
      "username": "a_nice_user"
    },
    "name": "hello_world_ci",
    "full_name": "a_nice_user/hello_world_ci",
    "private": false,
    "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
    "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
    "link": "",
    "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
    "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
    "default_branch": "main",
    "permissions": {
      "admin": true,
      "push": true,
      "pull": true
    },
    "object_format_name": "sha1",
  },
  "sender": {
    "id": 8765,
    "login": "a_nice_user",
    "full_name": "Nice User",
    "email": "a_nice_user@noreply.example.org",
    "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
    "html_url": "https://gitea.com/a_nice_user",
    "created": "2023-05-23T15:17:35+02:00",
    "visibility": "public",
    "username": "a_nice_user"
  },
  "review": null
}`

const HookPullRequestAddReviewRequestHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_review_request
`

const HookPullRequestAddReviewRequest = `{
  "action": "review_requested",
  "number": 7,
  "pull_request": {
    "id": 3779,
    "url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "number": 7,
    "user": {
      "id": 21,
      "login": "jony",
      "full_name": "Jony",
      "email": "jony@noreply.example.org",
      "avatar_url": "https://gitea.com/avatars/81027235e996f5e3ef6257152357b85d94171a2e",
      "html_url": "https://gitea.com/jony",
      "created": "2018-01-25T14:38:19+01:00",
      "visibility": "public",
      "username": "jony"
    },
    "title": "somepull",
    "body": "wow aaa new pulll body",
    "labels": [],
    "milestone": null,
    "assignees": null,
    "requested_reviewers": [
      {
        "id": 8765,
        "login": "a_nice_user",
        "full_name": "Nice User",
        "email": "a_nice_user@noreply.example.org",
        "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
        "html_url": "https://gitea.com/a_nice_user",
        "created": "2023-05-23T15:17:35+02:00",
        "visibility": "public",
        "username": "a_nice_user"
      }
    ],
    "state": "open",
    "additions": 1,
    "deletions": 0,
    "changed_files": 1,
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "diff_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.diff",
    "patch_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.patch",
    "base": {
      "label": "main",
      "ref": "main",
      "sha": "a40211c506550ebd79633d84e913dafa184c6d56",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "head": {
      "label": "jony-patch-1",
      "ref": "jony-patch-1",
      "sha": "07977177c2cd7d46bad37b8472a9d50e7acb9d1f",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "merge_base": "a40211c506550ebd79633d84e913dafa184c6d56",
    "due_date": null,
    "closed_at": null,
    "pin_order": 0
  },
  "requested_reviewer": {
    "id": 8765,
    "login": "a_nice_user",
    "full_name": "Nice User",
    "email": "a_nice_user@noreply.example.org",
    "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
    "html_url": "https://gitea.com/a_nice_user",
    "created": "2023-05-23T15:17:35+02:00",
    "visibility": "public",
    "username": "a_nice_user"
  },
  "repository": {
    "id": 1234,
    "owner": {
      "id": 8765,
      "login": "a_nice_user",
      "full_name": "Nice User",
      "email": "a_nice_user@me.mail",
      "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
      "html_url": "https://gitea.com/a_nice_user",
      "created": "2023-05-23T15:17:35+02:00",
      "visibility": "public",
      "username": "a_nice_user"
    },
    "name": "hello_world_ci",
    "full_name": "a_nice_user/hello_world_ci",
    "private": false,
    "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
    "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
    "link": "",
    "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
    "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
    "default_branch": "main",
    "permissions": {
      "admin": true,
      "push": true,
      "pull": true
    },
    "object_format_name": "sha1",
  },
  "sender": {
    "id": 8765,
    "login": "a_nice_user",
    "full_name": "Nice User",
    "email": "a_nice_user@noreply.example.org",
    "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
    "html_url": "https://gitea.com/a_nice_user",
    "created": "2023-05-23T15:17:35+02:00",
    "visibility": "public",
    "username": "a_nice_user"
  },
  "review": null
}`

const HookPullRequestAddLableHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_label
`

const HookPullRequestAddLable = `{
  "action": "label_updated",
  "number": 7,
  "pull_request": {
    "id": 3779,
    "url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "number": 7,
    "user": {
      "id": 21,
      "login": "jony",
      "full_name": "Jony",
      "email": "jony@noreply.example.org",
      "avatar_url": "https://gitea.com/avatars/81027235e996f5e3ef6257152357b85d94171a2e",
      "html_url": "https://gitea.com/jony",
      "created": "2018-01-25T14:38:19+01:00",
      "visibility": "public",
      "username": "jony"
    },
    "title": "somepull",
    "body": "wow aaa new pulll body",
    "labels": [
      {
        "id": 285,
        "name": "bug",
        "exclusive": false,
        "is_archived": false,
        "color": "ee0701",
        "description": "Something is not working",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/labels/285"
      },
      {
        "id": 297,
        "name": "help wanted",
        "exclusive": false,
        "is_archived": false,
        "color": "128a0c",
        "description": "Need some help",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/labels/297"
      }
    ],
    "milestone": null,
    "assignees": null,
    "requested_reviewers": [
      {
        "id": 8765,
        "login": "a_nice_user",
        "full_name": "Nice User",
        "email": "a_nice_user@noreply.example.org",
        "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
        "html_url": "https://gitea.com/a_nice_user",
        "created": "2023-05-23T15:17:35+02:00",
        "visibility": "public",
        "username": "a_nice_user"
      }
    ],
    "state": "open",
    "additions": 1,
    "deletions": 0,
    "changed_files": 1,
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "diff_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.diff",
    "patch_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.patch",
    "base": {
      "label": "main",
      "ref": "main",
      "sha": "a40211c506550ebd79633d84e913dafa184c6d56",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "head": {
      "label": "jony-patch-1",
      "ref": "jony-patch-1",
      "sha": "07977177c2cd7d46bad37b8472a9d50e7acb9d1f",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "merge_base": "a40211c506550ebd79633d84e913dafa184c6d56",
    "due_date": null,
    "closed_at": null,
    "pin_order": 0
  },
  "requested_reviewer": null,
  "repository": {
    "id": 1234,
    "owner": {
      "id": 8765,
      "login": "a_nice_user",
      "full_name": "Nice User",
      "email": "a_nice_user@me.mail",
      "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
      "html_url": "https://gitea.com/a_nice_user",
      "created": "2023-05-23T15:17:35+02:00",
      "visibility": "public",
      "username": "a_nice_user"
    },
    "name": "hello_world_ci",
    "full_name": "a_nice_user/hello_world_ci",
    "private": false,
    "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
    "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
    "link": "",
    "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
    "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
    "default_branch": "main",
    "permissions": {
      "admin": true,
      "push": true,
      "pull": true
    },
    "object_format_name": "sha1",
  },
  "sender": {
    "id": 8765,
    "login": "a_nice_user",
    "full_name": "Nice User",
    "email": "a_nice_user@noreply.example.org",
    "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
    "html_url": "https://gitea.com/a_nice_user",
    "created": "2023-05-23T15:17:35+02:00",
    "visibility": "public",
    "username": "a_nice_user"
  },
  "review": null
}`

const HookPullRequestChangeLableHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_label
`
const HookPullRequestChangeLable = `{
  "action": "label_updated",
  "number": 7,
  "pull_request": {
    "id": 3779,
    "url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "number": 7,
    "user": {
      "id": 21,
      "login": "jony",
      "full_name": "Jony",
      "email": "jony@noreply.example.org",
      "avatar_url": "https://gitea.com/avatars/81027235e996f5e3ef6257152357b85d94171a2e",
      "html_url": "https://gitea.com/jony",
      "created": "2018-01-25T14:38:19+01:00",
      "visibility": "public",
      "username": "jony"
    },
    "title": "somepull",
    "body": "wow aaa new pulll body",
    "labels": [
      {
        "id": 285,
        "name": "bug",
        "exclusive": false,
        "is_archived": false,
        "color": "ee0701",
        "description": "Something is not working",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/labels/285"
      }
    ],
    "milestone": null,
    "assignees": null,
    "requested_reviewers": [
      {
        "id": 8765,
        "login": "a_nice_user",
        "full_name": "Nice User",
        "email": "a_nice_user@noreply.example.org",
        "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
        "html_url": "https://gitea.com/a_nice_user",
        "created": "2023-05-23T15:17:35+02:00",
        "visibility": "public",
        "username": "a_nice_user"
      }
    ],
    "state": "open",
    "additions": 1,
    "deletions": 0,
    "changed_files": 1,
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "diff_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.diff",
    "patch_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.patch",
    "base": {
      "label": "main",
      "ref": "main",
      "sha": "a40211c506550ebd79633d84e913dafa184c6d56",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "head": {
      "label": "jony-patch-1",
      "ref": "jony-patch-1",
      "sha": "07977177c2cd7d46bad37b8472a9d50e7acb9d1f",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "merge_base": "a40211c506550ebd79633d84e913dafa184c6d56",
    "due_date": null,
    "closed_at": null,
    "pin_order": 0
  },
  "requested_reviewer": null,
  "repository": {
    "id": 1234,
    "owner": {
      "id": 8765,
      "login": "a_nice_user",
      "full_name": "Nice User",
      "email": "a_nice_user@me.mail",
      "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
      "html_url": "https://gitea.com/a_nice_user",
      "created": "2023-05-23T15:17:35+02:00",
      "visibility": "public",
      "username": "a_nice_user"
    },
    "name": "hello_world_ci",
    "full_name": "a_nice_user/hello_world_ci",
    "private": false,
    "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
    "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
    "link": "",
    "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
    "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
    "default_branch": "main",
    "permissions": {
      "admin": true,
      "push": true,
      "pull": true
    },
    "object_format_name": "sha1",
  },
  "sender": {
    "id": 8765,
    "login": "a_nice_user",
    "full_name": "Nice User",
    "email": "a_nice_user@noreply.example.org",
    "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
    "html_url": "https://gitea.com/a_nice_user",
    "created": "2023-05-23T15:17:35+02:00",
    "visibility": "public",
    "username": "a_nice_user"
  },
  "review": null
}`

const HookPullRequestRemoveLableHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_label
`
const HookPullRequestRemoveLable = `{
  "action": "label_cleared",
  "number": 7,
  "pull_request": {
    "id": 3779,
    "url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "number": 7,
    "user": {
      "id": 21,
      "login": "jony",
      "full_name": "Jony",
      "email": "jony@noreply.example.org",
      "avatar_url": "https://gitea.com/avatars/81027235e996f5e3ef6257152357b85d94171a2e",
      "html_url": "https://gitea.com/jony",
      "created": "2018-01-25T14:38:19+01:00",
      "visibility": "public",
      "username": "jony"
    },
    "title": "somepull",
    "body": "wow aaa new pulll body",
    "labels": [
      {
        "id": 285,
        "name": "bug",
        "exclusive": false,
        "is_archived": false,
        "color": "ee0701",
        "description": "Something is not working",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/labels/285"
      },
      {
        "id": 297,
        "name": "help wanted",
        "exclusive": false,
        "is_archived": false,
        "color": "128a0c",
        "description": "Need some help",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/labels/297"
      }
    ],
    "milestone": null,
    "assignees": null,
    "requested_reviewers": [
      {
        "id": 8765,
        "login": "a_nice_user",
        "full_name": "Nice User",
        "email": "a_nice_user@noreply.example.org",
        "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
        "html_url": "https://gitea.com/a_nice_user",
        "created": "2023-05-23T15:17:35+02:00",
        "visibility": "public",
        "username": "a_nice_user"
      }
    ],
    "state": "open",
    "additions": 1,
    "deletions": 0,
    "changed_files": 1,
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "diff_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.diff",
    "patch_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.patch",
    "base": {
      "label": "main",
      "ref": "main",
      "sha": "a40211c506550ebd79633d84e913dafa184c6d56",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "head": {
      "label": "jony-patch-1",
      "ref": "jony-patch-1",
      "sha": "07977177c2cd7d46bad37b8472a9d50e7acb9d1f",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "merge_base": "a40211c506550ebd79633d84e913dafa184c6d56",
    "due_date": null,
    "closed_at": null,
    "pin_order": 0
  },
  "requested_reviewer": null,
  "repository": {
    "id": 1234,
    "owner": {
      "id": 8765,
      "login": "a_nice_user",
      "full_name": "Nice User",
      "email": "a_nice_user@me.mail",
      "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
      "html_url": "https://gitea.com/a_nice_user",
      "created": "2023-05-23T15:17:35+02:00",
      "visibility": "public",
      "username": "a_nice_user"
    },
    "name": "hello_world_ci",
    "full_name": "a_nice_user/hello_world_ci",
    "private": false,
    "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
    "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
    "link": "",
    "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
    "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
    "default_branch": "main",
    "permissions": {
      "admin": true,
      "push": true,
      "pull": true
    },
    "object_format_name": "sha1",
  },
  "sender": {
    "id": 8765,
    "login": "a_nice_user",
    "full_name": "Nice User",
    "email": "a_nice_user@noreply.example.org",
    "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
    "html_url": "https://gitea.com/a_nice_user",
    "created": "2023-05-23T15:17:35+02:00",
    "visibility": "public",
    "username": "a_nice_user"
  },
  "review": null
}`

const HookPullRequestAddMileHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_milestone
`
const HookPullRequestAddMile = `{
  "action": "milestoned",
  "number": 7,
  "pull_request": {
    "id": 3779,
    "url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "number": 7,
    "user": {
      "id": 21,
      "login": "jony",
      "full_name": "Jony",
      "email": "jony@noreply.example.org",
      "avatar_url": "https://gitea.com/avatars/81027235e996f5e3ef6257152357b85d94171a2e",
      "html_url": "https://gitea.com/jony",
      "created": "2018-01-25T14:38:19+01:00",
      "visibility": "public",
      "username": "jony"
    },
    "title": "somepull",
    "body": "wow aaa new pulll body",
    "labels": [
      {
        "id": 285,
        "name": "bug",
        "exclusive": false,
        "is_archived": false,
        "color": "ee0701",
        "description": "Something is not working",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/labels/285"
      },
      {
        "id": 297,
        "name": "help wanted",
        "exclusive": false,
        "is_archived": false,
        "color": "128a0c",
        "description": "Need some help",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/labels/297"
      }
    ],
    "milestone": {
      "id": 277,
      "title": "new mile",
      "state": "open",
      "open_issues": 1,
      "closed_issues": 0,
      "closed_at": null,
      "due_on": null
    },
    "assignees": null,
    "requested_reviewers": [
      {
        "id": 8765,
        "login": "a_nice_user",
        "full_name": "Nice User",
        "email": "a_nice_user@noreply.example.org",
        "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
        "html_url": "https://gitea.com/a_nice_user",
        "created": "2023-05-23T15:17:35+02:00",
        "visibility": "public",
        "username": "a_nice_user"
      }
    ],
    "state": "open",
    "additions": 1,
    "deletions": 0,
    "changed_files": 1,
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "diff_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.diff",
    "patch_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.patch",
    "base": {
      "label": "main",
      "ref": "main",
      "sha": "a40211c506550ebd79633d84e913dafa184c6d56",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "head": {
      "label": "jony-patch-1",
      "ref": "jony-patch-1",
      "sha": "07977177c2cd7d46bad37b8472a9d50e7acb9d1f",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "merge_base": "a40211c506550ebd79633d84e913dafa184c6d56",
    "due_date": null,
    "closed_at": null,
    "pin_order": 0
  },
  "requested_reviewer": null,
  "repository": {
    "id": 1234,
    "owner": {
      "id": 8765,
      "login": "a_nice_user",
      "full_name": "Nice User",
      "email": "a_nice_user@me.mail",
      "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
      "html_url": "https://gitea.com/a_nice_user",
      "created": "2023-05-23T15:17:35+02:00",
      "visibility": "public",
      "username": "a_nice_user"
    },
    "name": "hello_world_ci",
    "full_name": "a_nice_user/hello_world_ci",
    "private": false,
    "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
    "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
    "link": "",
    "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
    "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
    "default_branch": "main",
    "permissions": {
      "admin": true,
      "push": true,
      "pull": true
    },
    "object_format_name": "sha1",
  },
  "sender": {
    "id": 8765,
    "login": "a_nice_user",
    "full_name": "Nice User",
    "email": "a_nice_user@noreply.example.org",
    "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
    "html_url": "https://gitea.com/a_nice_user",
    "created": "2023-05-23T15:17:35+02:00",
    "visibility": "public",
    "username": "a_nice_user"
  },
  "review": null
}`

const HookPullRequestChangeMileHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_milestone
`

const HookPullRequestChangeMile = `{
  "action": "milestoned",
  "number": 7,
  "pull_request": {
    "id": 3779,
    "url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "number": 7,
    "user": {
      "id": 21,
      "login": "jony",
      "full_name": "Jony",
      "email": "jony@noreply.example.org",
      "avatar_url": "https://gitea.com/avatars/81027235e996f5e3ef6257152357b85d94171a2e",
      "html_url": "https://gitea.com/jony",
      "created": "2018-01-25T14:38:19+01:00",
      "visibility": "public",
      "username": "jony"
    },
    "title": "somepull",
    "body": "wow aaa new pulll body",
    "labels": [
      {
        "id": 285,
        "name": "bug",
        "exclusive": false,
        "is_archived": false,
        "color": "ee0701",
        "description": "Something is not working",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/labels/285"
      },
      {
        "id": 297,
        "name": "help wanted",
        "exclusive": false,
        "is_archived": false,
        "color": "128a0c",
        "description": "Need some help",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/labels/297"
      }
    ],
    "milestone": {
      "id": 273,
      "title": "closed mile",
      "state": "closed",
      "open_issues": 1,
      "closed_issues": 0,
      "closed_at": "2025-05-28T03:13:46+02:00",
      "due_on": null
    },
    "assignees": null,
    "requested_reviewers": [
      {
        "id": 8765,
        "login": "a_nice_user",
        "full_name": "Nice User",
        "email": "a_nice_user@noreply.example.org",
        "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
        "html_url": "https://gitea.com/a_nice_user",
        "created": "2023-05-23T15:17:35+02:00",
        "visibility": "public",
        "username": "a_nice_user"
      }
    ],
    "state": "open",
    "additions": 1,
    "deletions": 0,
    "changed_files": 1,
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "diff_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.diff",
    "patch_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.patch",
    "base": {
      "label": "main",
      "ref": "main",
      "sha": "a40211c506550ebd79633d84e913dafa184c6d56",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "head": {
      "label": "jony-patch-1",
      "ref": "jony-patch-1",
      "sha": "07977177c2cd7d46bad37b8472a9d50e7acb9d1f",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "merge_base": "a40211c506550ebd79633d84e913dafa184c6d56",
    "due_date": null,
    "closed_at": null,
    "pin_order": 0
  },
  "requested_reviewer": null,
  "repository": {
    "id": 1234,
    "owner": {
      "id": 8765,
      "login": "a_nice_user",
      "full_name": "Nice User",
      "email": "a_nice_user@me.mail",
      "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
      "html_url": "https://gitea.com/a_nice_user",
      "created": "2023-05-23T15:17:35+02:00",
      "visibility": "public",
      "username": "a_nice_user"
    },
    "name": "hello_world_ci",
    "full_name": "a_nice_user/hello_world_ci",
    "private": false,
    "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
    "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
    "link": "",
    "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
    "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
    "default_branch": "main",
    "permissions": {
      "admin": true,
      "push": true,
      "pull": true
    },
    "object_format_name": "sha1",
  },
  "sender": {
    "id": 8765,
    "login": "a_nice_user",
    "full_name": "Nice User",
    "email": "a_nice_user@noreply.example.org",
    "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
    "html_url": "https://gitea.com/a_nice_user",
    "created": "2023-05-23T15:17:35+02:00",
    "visibility": "public",
    "username": "a_nice_user"
  },
  "review": null
}`

const HookPullRequestRemoveMileHeader = `
Request method: POST
Content-Type: application/json
X-Gitea-Event: pull_request
X-Gitea-Event-Type: pull_request_milestone
`
const HookPullRequestRemoveMile = `{
  "action": "demilestoned",
  "number": 7,
  "pull_request": {
    "id": 3779,
    "url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "number": 7,
    "user": {
      "id": 21,
      "login": "jony",
      "full_name": "Jony",
      "email": "jony@noreply.example.org",
      "avatar_url": "https://gitea.com/avatars/81027235e996f5e3ef6257152357b85d94171a2e",
      "html_url": "https://gitea.com/jony",
      "created": "2018-01-25T14:38:19+01:00",
      "visibility": "public",
      "username": "jony"
    },
    "title": "somepull",
    "body": "wow aaa new pulll body",
    "labels": [
      {
        "id": 285,
        "name": "bug",
        "exclusive": false,
        "is_archived": false,
        "color": "ee0701",
        "description": "Something is not working",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/labels/285"
      },
      {
        "id": 297,
        "name": "help wanted",
        "exclusive": false,
        "is_archived": false,
        "color": "128a0c",
        "description": "Need some help",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/labels/297"
      }
    ],
    "milestone": {
      "id": 273,
      "title": "closed mile",
      "state": "closed",
      "open_issues": 1,
      "closed_issues": 0,
      "closed_at": "2025-05-28T03:13:46+02:00",
      "due_on": null
    },
    "assignees": null,
    "requested_reviewers": [
      {
        "id": 8765,
        "login": "a_nice_user",
        "full_name": "Nice User",
        "email": "a_nice_user@noreply.example.org",
        "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
        "html_url": "https://gitea.com/a_nice_user",
        "created": "2023-05-23T15:17:35+02:00",
        "visibility": "public",
        "username": "a_nice_user"
      }
    ],
    "state": "open",
    "additions": 1,
    "deletions": 0,
    "changed_files": 1,
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7",
    "diff_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.diff",
    "patch_url": "https://gitea.com/a_nice_user/hello_world_ci/pulls/7.patch",
    "base": {
      "label": "main",
      "ref": "main",
      "sha": "a40211c506550ebd79633d84e913dafa184c6d56",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "head": {
      "label": "jony-patch-1",
      "ref": "jony-patch-1",
      "sha": "07977177c2cd7d46bad37b8472a9d50e7acb9d1f",
      "repo_id": 1234,
      "repo": {
        "id": 1234,
        "owner": {
          "id": 8765,
          "login": "a_nice_user",
          "full_name": "Nice User",
          "email": "a_nice_user@me.mail",
          "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
          "html_url": "https://gitea.com/a_nice_user",
          "created": "2023-05-23T15:17:35+02:00",
          "visibility": "public",
          "username": "a_nice_user"
        },
        "name": "hello_world_ci",
        "full_name": "a_nice_user/hello_world_ci",
        "private": false,
        "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
        "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
        "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
        "link": "",
        "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
        "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
        "default_branch": "main",
        "permissions": {
          "admin": false,
          "push": false,
          "pull": true
        },
        "object_format_name": "sha1",
      }
    },
    "merge_base": "a40211c506550ebd79633d84e913dafa184c6d56",
    "due_date": null,
    "closed_at": null,
    "pin_order": 0
  },
  "requested_reviewer": null,
  "repository": {
    "id": 1234,
    "owner": {
      "id": 8765,
      "login": "a_nice_user",
      "full_name": "Nice User",
      "email": "a_nice_user@me.mail",
      "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
      "html_url": "https://gitea.com/a_nice_user",
      "created": "2023-05-23T15:17:35+02:00",
      "visibility": "public",
      "username": "a_nice_user"
    },
    "name": "hello_world_ci",
    "full_name": "a_nice_user/hello_world_ci",
    "private": false,
    "languages_url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci/languages",
    "html_url": "https://gitea.com/a_nice_user/hello_world_ci",
    "url": "https://gitea.com/api/v1/repos/a_nice_user/hello_world_ci",
    "link": "",
    "ssh_url": "ssh://git@gitea.rt4u.de:3232/a_nice_user/hello_world_ci.git",
    "clone_url": "https://gitea.com/a_nice_user/hello_world_ci.git",
    "default_branch": "main",
    "permissions": {
      "admin": true,
      "push": true,
      "pull": true
    },
    "object_format_name": "sha1",
  },
  "sender": {
    "id": 8765,
    "login": "a_nice_user",
    "full_name": "Nice User",
    "email": "a_nice_user@noreply.example.org",
    "avatar_url": "https://gitea.com/avatars/ae32f5573b27f9840942a522d59032b104a2dd15",
    "html_url": "https://gitea.com/a_nice_user",
    "created": "2023-05-23T15:17:35+02:00",
    "visibility": "public",
    "username": "a_nice_user"
  },
  "review": null
}`

const HookRelease = `
{
  "action": "published",
  "release": {
    "id": 48,
    "tag_name": "0.0.5",
    "target_commitish": "main",
    "name": "Version 0.0.5",
    "body": "",
    "url": "https://git.xxx/api/v1/repos/anbraten/demo/releases/48",
    "html_url": "https://git.xxx/anbraten/demo/releases/tag/0.0.5",
    "tarball_url": "https://git.xxx/anbraten/demo/archive/0.0.5.tar.gz",
    "zipball_url": "https://git.xxx/anbraten/demo/archive/0.0.5.zip",
    "draft": false,
    "prerelease": false,
    "created_at": "2022-02-09T20:23:05Z",
    "published_at": "2022-02-09T20:23:05Z",
    "author": {"id":1,"login":"anbraten","full_name":"Anton Bracke","email":"anbraten@noreply.xxx","avatar_url":"https://git.xxx/user/avatar/anbraten/-1","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2018-03-21T10:04:48Z","restricted":false,"active":false,"prohibit_login":false,"location":"world","website":"https://xxx","description":"","visibility":"public","followers_count":1,"following_count":1,"starred_repos_count":1,"username":"anbraten"},
    "assets": []
  },
  "repository": {
    "id": 77,
    "owner": {"id":1,"login":"anbraten","full_name":"Anton Bracke","email":"anbraten@noreply.xxx","avatar_url":"https://git.xxx/user/avatar/anbraten/-1","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2018-03-21T10:04:48Z","restricted":false,"active":false,"prohibit_login":false,"location":"world","website":"https://xxx","description":"","visibility":"public","followers_count":1,"following_count":1,"starred_repos_count":1,"username":"anbraten"},
    "name": "demo",
    "full_name": "anbraten/demo",
    "description": "",
    "empty": false,
    "private": true,
    "fork": false,
    "template": false,
    "parent": null,
    "mirror": false,
    "size": 59,
    "html_url": "https://git.xxx/anbraten/demo",
    "ssh_url": "ssh://git@git.xxx:22/anbraten/demo.git",
    "clone_url": "https://git.xxx/anbraten/demo.git",
    "original_url": "",
    "website": "",
    "stars_count": 0,
    "forks_count": 1,
    "watchers_count": 1,
    "open_issues_count": 2,
    "open_pr_counter": 2,
    "release_counter": 4,
    "default_branch": "main",
    "archived": false,
    "created_at": "2021-08-30T20:54:13Z",
    "updated_at": "2022-01-09T01:29:23Z",
    "permissions": {
      "admin": true,
      "push": true,
      "pull": true
    },
    "has_issues": true,
    "internal_tracker": {
      "enable_time_tracker": true,
      "allow_only_contributors_to_track_time": true,
      "enable_issue_dependencies": true
    },
    "has_wiki": false,
    "has_pull_requests": true,
    "has_projects": true,
    "ignore_whitespace_conflicts": false,
    "allow_merge_commits": true,
    "allow_rebase": true,
    "allow_rebase_explicit": true,
    "allow_squash_merge": true,
    "default_merge_style": "squash",
    "avatar_url": "",
    "internal": false,
    "mirror_interval": ""
  },
  "sender": {"id":1,"login":"anbraten","full_name":"Anbraten","email":"anbraten@noreply.xxx","avatar_url":"https://git.xxx/user/avatar/anbraten/-1","language":"","is_admin":false,"last_login":"0001-01-01T00:00:00Z","created":"2018-03-21T10:04:48Z","restricted":false,"active":false,"prohibit_login":false,"location":"World","website":"https://xxx","description":"","visibility":"public","followers_count":1,"following_count":1,"starred_repos_count":1,"username":"anbraten"}
}
`

//go:embed HookRelease.json
var HookRelease string
