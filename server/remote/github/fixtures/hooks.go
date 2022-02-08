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

// HookPush is a sample push hook.
// https://developer.github.com/v3/activity/events/types/#pushevent
const HookPush = `{
  "ref": "refs/heads/master",
  "before": "2f780193b136b72bfea4eeb640786a8c4450c7a2",
  "after": "366701fde727cb7a9e7f21eb88264f59f6f9b89c",
  "repository": {
    "id": 179344069,
    "node_id": "MDEwOlJlcG9zaXRvcnkxNzkzNDQwNjk=",
    "name": "woodpecker",
    "full_name": "woodpecker-ci/woodpecker",
    "private": false,
    "owner": {
      "name": "woodpecker-ci",
      "email": null,
      "login": "woodpecker-ci",
      "id": 84780935,
      "node_id": "MDEyOk9yZ2FuaXphdGlvbjg0NzgwOTM1",
      "avatar_url": "https://avatars.githubusercontent.com/u/84780935?v=4",
      "gravatar_id": "",
      "url": "https://api.github.com/users/woodpecker-ci",
      "html_url": "https://github.com/woodpecker-ci",
      "followers_url": "https://api.github.com/users/woodpecker-ci/followers",
      "following_url": "https://api.github.com/users/woodpecker-ci/following{/other_user}",
      "gists_url": "https://api.github.com/users/woodpecker-ci/gists{/gist_id}",
      "starred_url": "https://api.github.com/users/woodpecker-ci/starred{/owner}{/repo}",
      "subscriptions_url": "https://api.github.com/users/woodpecker-ci/subscriptions",
      "organizations_url": "https://api.github.com/users/woodpecker-ci/orgs",
      "repos_url": "https://api.github.com/users/woodpecker-ci/repos",
      "events_url": "https://api.github.com/users/woodpecker-ci/events{/privacy}",
      "received_events_url": "https://api.github.com/users/woodpecker-ci/received_events",
      "type": "Organization",
      "site_admin": false
    },
    "html_url": "https://github.com/woodpecker-ci/woodpecker",
    "description": "Woodpecker is a community fork of the Drone CI system.",
    "fork": false,
    "url": "https://github.com/woodpecker-ci/woodpecker",
    "forks_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/forks",
    "keys_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/keys{/key_id}",
    "collaborators_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/collaborators{/collaborator}",
    "teams_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/teams",
    "hooks_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/hooks",
    "issue_events_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/issues/events{/number}",
    "events_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/events",
    "assignees_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/assignees{/user}",
    "branches_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/branches{/branch}",
    "tags_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/tags",
    "blobs_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/git/blobs{/sha}",
    "git_tags_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/git/tags{/sha}",
    "git_refs_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/git/refs{/sha}",
    "trees_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/git/trees{/sha}",
    "statuses_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/statuses/{sha}",
    "languages_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/languages",
    "stargazers_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/stargazers",
    "contributors_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/contributors",
    "subscribers_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/subscribers",
    "subscription_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/subscription",
    "commits_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/commits{/sha}",
    "git_commits_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/git/commits{/sha}",
    "comments_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/comments{/number}",
    "issue_comment_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/issues/comments{/number}",
    "contents_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/contents/{+path}",
    "compare_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/compare/{base}...{head}",
    "merges_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/merges",
    "archive_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/{archive_format}{/ref}",
    "downloads_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/downloads",
    "issues_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/issues{/number}",
    "pulls_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/pulls{/number}",
    "milestones_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/milestones{/number}",
    "notifications_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/notifications{?since,all,participating}",
    "labels_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/labels{/name}",
    "releases_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/releases{/id}",
    "deployments_url": "https://api.github.com/repos/woodpecker-ci/woodpecker/deployments",
    "created_at": 1554314798,
    "updated_at": "2022-01-16T20:19:33Z",
    "pushed_at": 1642370257,
    "git_url": "git://github.com/woodpecker-ci/woodpecker.git",
    "ssh_url": "git@github.com:woodpecker-ci/woodpecker.git",
    "clone_url": "https://github.com/woodpecker-ci/woodpecker.git",
    "svn_url": "https://github.com/woodpecker-ci/woodpecker",
    "homepage": "https://woodpecker-ci.org",
    "size": 81324,
    "stargazers_count": 659,
    "watchers_count": 659,
    "language": "Go",
    "has_issues": true,
    "has_projects": false,
    "has_downloads": true,
    "has_wiki": false,
    "has_pages": false,
    "forks_count": 84,
    "mirror_url": null,
    "archived": false,
    "disabled": false,
    "open_issues_count": 123,
    "license": {
      "key": "apache-2.0",
      "name": "Apache License 2.0",
      "spdx_id": "Apache-2.0",
      "url": "https://api.github.com/licenses/apache-2.0",
      "node_id": "MDc6TGljZW5zZTI="
    },
    "allow_forking": true,
    "is_template": false,
    "topics": [
      "ci",
      "devops",
      "docker",
      "hacktoberfest",
      "hacktoberfest2021",
      "woodpeckerci"
    ],
    "visibility": "public",
    "forks": 84,
    "open_issues": 123,
    "watchers": 659,
    "default_branch": "master",
    "stargazers": 659,
    "master_branch": "master",
    "organization": "woodpecker-ci"
  },
  "pusher": {
    "name": "6543",
    "email": "noreply@6543.de"
  },
  "organization": {
    "login": "woodpecker-ci",
    "id": 84780935,
    "node_id": "MDEyOk9yZ2FuaXphdGlvbjg0NzgwOTM1",
    "url": "https://api.github.com/orgs/woodpecker-ci",
    "repos_url": "https://api.github.com/orgs/woodpecker-ci/repos",
    "events_url": "https://api.github.com/orgs/woodpecker-ci/events",
    "hooks_url": "https://api.github.com/orgs/woodpecker-ci/hooks",
    "issues_url": "https://api.github.com/orgs/woodpecker-ci/issues",
    "members_url": "https://api.github.com/orgs/woodpecker-ci/members{/member}",
    "public_members_url": "https://api.github.com/orgs/woodpecker-ci/public_members{/member}",
    "avatar_url": "https://avatars.githubusercontent.com/u/84780935?v=4",
    "description": "Woodpecker is a community fork of the Drone CI system."
  },
  "sender": {
    "login": "6543",
    "id": 24977596,
    "node_id": "MDQ6VXNlcjI0OTc3NTk2",
    "avatar_url": "https://avatars.githubusercontent.com/u/24977596?v=4",
    "gravatar_id": "",
    "url": "https://api.github.com/users/6543",
    "html_url": "https://github.com/6543",
    "followers_url": "https://api.github.com/users/6543/followers",
    "following_url": "https://api.github.com/users/6543/following{/other_user}",
    "gists_url": "https://api.github.com/users/6543/gists{/gist_id}",
    "starred_url": "https://api.github.com/users/6543/starred{/owner}{/repo}",
    "subscriptions_url": "https://api.github.com/users/6543/subscriptions",
    "organizations_url": "https://api.github.com/users/6543/orgs",
    "repos_url": "https://api.github.com/users/6543/repos",
    "events_url": "https://api.github.com/users/6543/events{/privacy}",
    "received_events_url": "https://api.github.com/users/6543/received_events",
    "type": "User",
    "site_admin": false
  },
  "created": false,
  "deleted": false,
  "forced": false,
  "base_ref": null,
  "compare": "https://github.com/woodpecker-ci/woodpecker/compare/2f780193b136...366701fde727",
  "commits": [
    {
      "id": "366701fde727cb7a9e7f21eb88264f59f6f9b89c",
      "tree_id": "638e046f1e1e15dbed1ddf40f9471bf1af4d64ce",
      "distinct": true,
      "message": "Fix multiline secrets replacer (#700)\n\n* Fix multiline secrets replacer\r\n\r\n* Add tests",
      "timestamp": "2022-01-16T22:57:37+01:00",
      "url": "https://github.com/woodpecker-ci/woodpecker/commit/366701fde727cb7a9e7f21eb88264f59f6f9b89c",
      "author": {
        "name": "Philipp",
        "email": "noreply@philipp.xzy",
        "username": "nupplaphil"
      },
      "committer": {
        "name": "GitHub",
        "email": "noreply@github.com",
        "username": "web-flow"
      },
      "added": [

      ],
      "removed": [

      ],
      "modified": [
        "pipeline/shared/replace_secrets.go",
        "pipeline/shared/replace_secrets_test.go"
      ]
    }
  ],
  "head_commit": {
    "id": "366701fde727cb7a9e7f21eb88264f59f6f9b89c",
    "tree_id": "638e046f1e1e15dbed1ddf40f9471bf1af4d64ce",
    "distinct": true,
    "message": "Fix multiline secrets replacer (#700)\n\n* Fix multiline secrets replacer\r\n\r\n* Add tests",
    "timestamp": "2022-01-16T22:57:37+01:00",
    "url": "https://github.com/woodpecker-ci/woodpecker/commit/366701fde727cb7a9e7f21eb88264f59f6f9b89c",
    "author": {
      "name": "Philipp",
      "email": "admin@philipp.info",
      "username": "nupplaphil"
    },
    "committer": {
      "name": "GitHub",
      "email": "noreply@github.com",
      "username": "web-flow"
    },
    "added": [

    ],
    "removed": [

    ],
    "modified": [
      "pipeline/shared/replace_secrets.go",
      "pipeline/shared/replace_secrets_test.go"
    ]
  }
}`

// HookPushDeleted is a sample push hook that is marked as deleted, and is expected to be ignored.
const HookPushDeleted = `
{
  "deleted": true
}
`

// HookPullRequest is a sample hook pull request
// https://developer.github.com/v3/activity/events/types/#pullrequestevent
const HookPullRequest = `
{
  "action": "opened",
  "number": 1,
  "pull_request": {
    "url": "https://api.github.com/repos/baxterthehacker/public-repo/pulls/1",
    "html_url": "https://github.com/baxterthehacker/public-repo/pull/1",
    "number": 1,
    "state": "open",
    "title": "Update the README with new information",
    "user": {
      "login": "baxterthehacker",
      "avatar_url": "https://avatars.githubusercontent.com/u/6752317?v=3"
    },
    "base": {
      "label": "baxterthehacker:master",
      "ref": "master",
      "sha": "9353195a19e45482665306e466c832c46560532d"
    },
    "head": {
      "label": "baxterthehacker:changes",
      "ref": "changes",
      "sha": "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c"
    }
  },
  "repository": {
    "id": 35129377,
    "name": "public-repo",
    "full_name": "baxterthehacker/public-repo",
    "owner": {
      "login": "baxterthehacker",
      "avatar_url": "https://avatars.githubusercontent.com/u/6752317?v=3"
    },
    "private": true,
    "html_url": "https://github.com/baxterthehacker/public-repo",
    "clone_url": "https://github.com/baxterthehacker/public-repo.git",
    "default_branch": "master"
  },
  "sender": {
    "login": "octocat",
    "avatar_url": "https://avatars.githubusercontent.com/u/6752317?v=3"
  }
}
`

// HookPullRequestInvalidAction is a sample hook pull request that has an
// action not equal to synchrize or opened, and is expected to be ignored.
const HookPullRequestInvalidAction = `
{
  "action": "reopened",
  "number": 1
}
`

// HookPullRequestInvalidState is a sample hook pull request that has a state
// not equal to open, and is expected to be ignored.
const HookPullRequestInvalidState = `
{
  "action": "synchronize",
  "pull_request": {
    "number": 1,
    "state": "closed"
  }
}
`

// HookPush is a sample deployment hook.
// https://developer.github.com/v3/activity/events/types/#deploymentevent
const HookDeploy = `
{
  "deployment": {
    "url": "https://api.github.com/repos/baxterthehacker/public-repo/deployments/710692",
    "id": 710692,
    "sha": "9049f1265b7d61be4a8904a9a27120d2064dab3b",
    "ref": "master",
    "task": "deploy",
    "payload": {
    },
    "environment": "production",
    "description": null,
    "creator": {
      "login": "baxterthehacker",
      "avatar_url": "https://avatars.githubusercontent.com/u/6752317?v=3"
    }
  },
  "repository": {
    "id": 35129377,
    "name": "public-repo",
    "full_name": "baxterthehacker/public-repo",
    "owner": {
      "login": "baxterthehacker",
      "avatar_url": "https://avatars.githubusercontent.com/u/6752317?v=3"
    },
    "private": true,
    "html_url": "https://github.com/baxterthehacker/public-repo",
    "clone_url": "https://github.com/baxterthehacker/public-repo.git",
    "default_branch": "master"
  },
  "sender": {
    "login": "baxterthehacker",
    "avatar_url": "https://avatars.githubusercontent.com/u/6752317?v=3"
  }
}
`

// HookPush is a sample release hook.
// https://developer.github.com/v3/activity/events/types/#release
const HookRelease = `
{
  "action": "published",
  "release": {
    "url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/releases/2",
    "assets_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/releases/2/assets",
    "upload_url": "https://octocoders.github.io/api/uploads/repos/Codertocat/Hello-World/releases/2/assets{?name,label}",
    "html_url": "https://octocoders.github.io/Codertocat/Hello-World/releases/tag/0.0.1",
    "id": 2,
    "node_id": "MDc6UmVsZWFzZTI=",
    "tag_name": "0.0.1",
    "target_commitish": "master",
    "name": null,
    "draft": false,
    "author": {
      "login": "Codertocat",
      "id": 4,
      "node_id": "MDQ6VXNlcjQ=",
      "avatar_url": "https://octocoders.github.io/avatars/u/4?",
      "gravatar_id": "",
      "url": "https://octocoders.github.io/api/v3/users/Codertocat",
      "html_url": "https://octocoders.github.io/Codertocat",
      "followers_url": "https://octocoders.github.io/api/v3/users/Codertocat/followers",
      "following_url": "https://octocoders.github.io/api/v3/users/Codertocat/following{/other_user}",
      "gists_url": "https://octocoders.github.io/api/v3/users/Codertocat/gists{/gist_id}",
      "starred_url": "https://octocoders.github.io/api/v3/users/Codertocat/starred{/owner}{/repo}",
      "subscriptions_url": "https://octocoders.github.io/api/v3/users/Codertocat/subscriptions",
      "organizations_url": "https://octocoders.github.io/api/v3/users/Codertocat/orgs",
      "repos_url": "https://octocoders.github.io/api/v3/users/Codertocat/repos",
      "events_url": "https://octocoders.github.io/api/v3/users/Codertocat/events{/privacy}",
      "received_events_url": "https://octocoders.github.io/api/v3/users/Codertocat/received_events",
      "type": "User",
      "site_admin": false
    },
    "prerelease": false,
    "created_at": "2019-05-15T19:37:08Z",
    "published_at": "2019-05-15T19:38:20Z",
    "assets": [],
    "tarball_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/tarball/0.0.1",
    "zipball_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/zipball/0.0.1",
    "body": null
  },
  "repository": {
    "id": 118,
    "node_id": "MDEwOlJlcG9zaXRvcnkxMTg=",
    "name": "Hello-World",
    "full_name": "Codertocat/Hello-World",
    "private": false,
    "owner": {
      "login": "Codertocat",
      "id": 4,
      "node_id": "MDQ6VXNlcjQ=",
      "avatar_url": "https://octocoders.github.io/avatars/u/4?",
      "gravatar_id": "",
      "url": "https://octocoders.github.io/api/v3/users/Codertocat",
      "html_url": "https://octocoders.github.io/Codertocat",
      "followers_url": "https://octocoders.github.io/api/v3/users/Codertocat/followers",
      "following_url": "https://octocoders.github.io/api/v3/users/Codertocat/following{/other_user}",
      "gists_url": "https://octocoders.github.io/api/v3/users/Codertocat/gists{/gist_id}",
      "starred_url": "https://octocoders.github.io/api/v3/users/Codertocat/starred{/owner}{/repo}",
      "subscriptions_url": "https://octocoders.github.io/api/v3/users/Codertocat/subscriptions",
      "organizations_url": "https://octocoders.github.io/api/v3/users/Codertocat/orgs",
      "repos_url": "https://octocoders.github.io/api/v3/users/Codertocat/repos",
      "events_url": "https://octocoders.github.io/api/v3/users/Codertocat/events{/privacy}",
      "received_events_url": "https://octocoders.github.io/api/v3/users/Codertocat/received_events",
      "type": "User",
      "site_admin": false
    },
    "html_url": "https://octocoders.github.io/Codertocat/Hello-World",
    "description": null,
    "fork": false,
    "url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World",
    "forks_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/forks",
    "keys_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/keys{/key_id}",
    "collaborators_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/collaborators{/collaborator}",
    "teams_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/teams",
    "hooks_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/hooks",
    "issue_events_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/issues/events{/number}",
    "events_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/events",
    "assignees_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/assignees{/user}",
    "branches_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/branches{/branch}",
    "tags_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/tags",
    "blobs_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/git/blobs{/sha}",
    "git_tags_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/git/tags{/sha}",
    "git_refs_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/git/refs{/sha}",
    "trees_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/git/trees{/sha}",
    "statuses_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/statuses/{sha}",
    "languages_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/languages",
    "stargazers_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/stargazers",
    "contributors_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/contributors",
    "subscribers_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/subscribers",
    "subscription_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/subscription",
    "commits_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/commits{/sha}",
    "git_commits_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/git/commits{/sha}",
    "comments_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/comments{/number}",
    "issue_comment_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/issues/comments{/number}",
    "contents_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/contents/{+path}",
    "compare_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/compare/{base}...{head}",
    "merges_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/merges",
    "archive_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/{archive_format}{/ref}",
    "downloads_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/downloads",
    "issues_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/issues{/number}",
    "pulls_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/pulls{/number}",
    "milestones_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/milestones{/number}",
    "notifications_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/notifications{?since,all,participating}",
    "labels_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/labels{/name}",
    "releases_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/releases{/id}",
    "deployments_url": "https://octocoders.github.io/api/v3/repos/Codertocat/Hello-World/deployments",
    "created_at": "2019-05-15T19:37:07Z",
    "updated_at": "2019-05-15T19:38:15Z",
    "pushed_at": "2019-05-15T19:38:19Z",
    "git_url": "git://octocoders.github.io/Codertocat/Hello-World.git",
    "ssh_url": "git@octocoders.github.io:Codertocat/Hello-World.git",
    "clone_url": "https://octocoders.github.io/Codertocat/Hello-World.git",
    "svn_url": "https://octocoders.github.io/Codertocat/Hello-World",
    "homepage": null,
    "size": 0,
    "stargazers_count": 0,
    "watchers_count": 0,
    "language": "Ruby",
    "has_issues": true,
    "has_projects": true,
    "has_downloads": true,
    "has_wiki": true,
    "has_pages": true,
    "forks_count": 1,
    "mirror_url": null,
    "archived": false,
    "disabled": false,
    "open_issues_count": 2,
    "license": null,
    "forks": 1,
    "open_issues": 2,
    "watchers": 0,
    "default_branch": "master"
  },
  "enterprise": {
    "id": 1,
    "slug": "github",
    "name": "GitHub",
    "node_id": "MDg6QnVzaW5lc3Mx",
    "avatar_url": "https://octocoders.github.io/avatars/b/1?",
    "description": null,
    "website_url": null,
    "html_url": "https://octocoders.github.io/businesses/github",
    "created_at": "2019-05-14T19:31:12Z",
    "updated_at": "2019-05-14T19:31:12Z"
  },
  "sender": {
    "login": "Codertocat",
    "id": 4,
    "node_id": "MDQ6VXNlcjQ=",
    "avatar_url": "https://octocoders.github.io/avatars/u/4?",
    "gravatar_id": "",
    "url": "https://octocoders.github.io/api/v3/users/Codertocat",
    "html_url": "https://octocoders.github.io/Codertocat",
    "followers_url": "https://octocoders.github.io/api/v3/users/Codertocat/followers",
    "following_url": "https://octocoders.github.io/api/v3/users/Codertocat/following{/other_user}",
    "gists_url": "https://octocoders.github.io/api/v3/users/Codertocat/gists{/gist_id}",
    "starred_url": "https://octocoders.github.io/api/v3/users/Codertocat/starred{/owner}{/repo}",
    "subscriptions_url": "https://octocoders.github.io/api/v3/users/Codertocat/subscriptions",
    "organizations_url": "https://octocoders.github.io/api/v3/users/Codertocat/orgs",
    "repos_url": "https://octocoders.github.io/api/v3/users/Codertocat/repos",
    "events_url": "https://octocoders.github.io/api/v3/users/Codertocat/events{/privacy}",
    "received_events_url": "https://octocoders.github.io/api/v3/users/Codertocat/received_events",
    "type": "User",
    "site_admin": false
  },
  "installation": {
    "id": 5,
    "node_id": "MDIzOkludGVncmF0aW9uSW5zdGFsbGF0aW9uNQ=="
  }
}

`
