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

// HookPush is a sample Forgejo push hook
const HookPush = `
{
  "ref": "refs/heads/master",
  "before": "4b2626259b5a97b6b4eab5e6cca66adb986b672b",
  "after": "ef98532add3b2feb7a137426bba1248724367df5",
  "compare_url": "http://forgejo.golang.org/gordon/hello-world/compare/4b2626259b5a97b6b4eab5e6cca66adb986b672b...ef98532add3b2feb7a137426bba1248724367df5",
  "commits": [
    {
      "id": "ef98532add3b2feb7a137426bba1248724367df5",
      "message": "bump\n",
      "url": "http://forgejo.golang.org/gordon/hello-world/commit/ef98532add3b2feb7a137426bba1248724367df5",
      "author": {
        "name": "Gordon the Gopher",
        "email": "gordon@golang.org",
        "username": "gordon"
      },
      "added": ["CHANGELOG.md"],
      "removed": [],
      "modified": ["app/controller/application.rb"]
    }
  ],
  "repository": {
    "id": 1,
    "name": "hello-world",
    "full_name": "gordon/hello-world",
    "html_url": "http://forgejo.golang.org/gordon/hello-world",
    "ssh_url": "git@forgejo.golang.org:gordon/hello-world.git",
    "clone_url": "http://forgejo.golang.org/gordon/hello-world.git",
    "description": "",
    "website": "",
    "watchers": 1,
    "owner": {
      "name": "gordon",
      "email": "gordon@golang.org",
      "login": "gordon",
      "username": "gordon"
    },
    "private": true
  },
  "pusher": {
    "name": "gordon",
    "email": "gordon@golang.org",
    "username": "gordon",
    "login": "gordon"
  },
  "sender": {
    "login": "gordon",
    "id": 1,
    "username": "gordon",
    "email": "gordon@golang.org",
    "avatar_url": "http://forgejo.golang.org///1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87"
  }
}
`

// HookPushBranch is a sample Forgejo push hook where a new branch was created from an existing commit
const HookPushBranch = `
{
  "ref": "refs/heads/fdsafdsa",
  "before": "0000000000000000000000000000000000000000",
  "after": "28c3613ae62640216bea5e7dc71aa65356e4298b",
  "compare_url": "https://codeberg.org/meisam/woodpecktester/compare/master...28c3613ae62640216bea5e7dc71aa65356e4298b",
  "commits": [],
  "head_commit": {
    "id": "28c3613ae62640216bea5e7dc71aa65356e4298b",
    "message": "Delete '.woodpecker/.check.yml'\n",
    "url": "https://codeberg.org/meisam/woodpecktester/commit/28c3613ae62640216bea5e7dc71aa65356e4298b",
    "author": {
      "name": "meisam",
      "email": "meisam@noreply.codeberg.org",
      "username": "meisam"
    },
    "committer": {
      "name": "meisam",
      "email": "meisam@noreply.codeberg.org",
      "username": "meisam"
    },
    "verification": null,
    "timestamp": "2022-07-12T21:09:27+02:00",
    "added": [],
    "removed": [
      ".woodpecker/.check.yml"
    ],
    "modified": []
  },
  "repository": {
    "id": 50820,
    "owner": {
      "id": 14844,
      "login": "meisam",
      "full_name": "",
      "email": "meisam@noreply.codeberg.org",
      "avatar_url": "https://codeberg.org/avatars/96512da76a14cf44e0bb32d1640e878e",
      "language": "",
      "is_admin": false,
      "last_login": "0001-01-01T00:00:00Z",
      "created": "2020-10-08T11:19:12+02:00",
      "restricted": false,
      "active": false,
      "prohibit_login": false,
      "location": "",
      "website": "",
      "description": "Materials engineer, physics enthusiast, large collection of the bad programming habits, always happy to fix the old ones and make new mistakes!",
      "visibility": "public",
      "followers_count": 0,
      "following_count": 0,
      "starred_repos_count": 0,
      "username": "meisam"
    },
    "name": "woodpecktester",
    "full_name": "meisam/woodpecktester",
    "description": "Just for testing the Woodpecker CI and reporting bugs",
    "empty": false,
    "private": false,
    "fork": false,
    "template": false,
    "parent": null,
    "mirror": false,
    "size": 367,
    "language": "",
    "languages_url": "https://codeberg.org/api/v1/repos/meisam/woodpecktester/languages",
    "html_url": "https://codeberg.org/meisam/woodpecktester",
    "ssh_url": "git@codeberg.org:meisam/woodpecktester.git",
    "clone_url": "https://codeberg.org/meisam/woodpecktester.git",
    "original_url": "",
    "website": "",
    "stars_count": 0,
    "forks_count": 0,
    "watchers_count": 1,
    "open_issues_count": 0,
    "open_pr_counter": 0,
    "release_counter": 0,
    "default_branch": "master",
    "archived": false,
    "created_at": "2022-07-04T00:34:39+02:00",
    "updated_at": "2022-07-24T20:31:29+02:00",
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
    "has_wiki": true,
    "has_pull_requests": true,
    "has_projects": true,
    "ignore_whitespace_conflicts": false,
    "allow_merge_commits": true,
    "allow_rebase": true,
    "allow_rebase_explicit": true,
    "allow_squash_merge": true,
    "default_merge_style": "merge",
    "avatar_url": "",
    "internal": false,
    "mirror_interval": "",
    "mirror_updated": "0001-01-01T00:00:00Z",
    "repo_transfer": null
  },
  "pusher": {
    "id": 2628,
    "login": "6543",
    "full_name": "",
    "email": "6543@obermui.de",
    "avatar_url": "https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173",
    "language": "",
    "is_admin": false,
    "last_login": "0001-01-01T00:00:00Z",
    "created": "2019-10-12T05:05:49+02:00",
    "restricted": false,
    "active": false,
    "prohibit_login": false,
    "location": "",
    "visibility": "public",
    "followers_count": 22,
    "following_count": 16,
    "starred_repos_count": 55,
    "username": "6543"
  },
  "sender": {
    "id": 2628,
    "login": "6543",
    "full_name": "",
    "email": "6543@obermui.de",
    "avatar_url": "https://codeberg.org/avatars/09a234c768cb9bca78f6b2f82d6af173",
    "language": "",
    "is_admin": false,
    "last_login": "0001-01-01T00:00:00Z",
    "created": "2019-10-12T05:05:49+02:00",
    "restricted": false,
    "active": false,
    "prohibit_login": false,
    "visibility": "public",
    "followers_count": 22,
    "following_count": 16,
    "starred_repos_count": 55,
    "username": "6543"
  }
}
`

// HookPushTag is a sample Forgejo tag hook
const HookPushTag = `{
  "sha": "ef98532add3b2feb7a137426bba1248724367df5",
  "secret": "l26Un7G7HXogLAvsyf2hOA4EMARSTsR3",
  "ref": "v1.0.0",
  "ref_type": "tag",
  "repository": {
    "id": 1,
    "owner": {
      "id": 1,
      "username": "gordon",
      "login": "gordon",
      "full_name": "Gordon the Gopher",
      "email": "gordon@golang.org",
      "avatar_url": "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87"
    },
    "name": "hello-world",
    "full_name": "gordon/hello-world",
    "description": "",
    "private": true,
    "fork": false,
    "html_url": "http://forgejo.golang.org/gordon/hello-world",
    "ssh_url": "git@forgejo.golang.org:gordon/hello-world.git",
    "clone_url": "http://forgejo.golang.org/gordon/hello-world.git",
    "default_branch": "master",
    "created_at": "2015-10-22T19:32:44Z",
    "updated_at": "2016-11-24T13:37:16Z"
  },
  "sender": {
    "id": 1,
    "username": "gordon",
    "login": "gordon",
    "full_name": "Gordon the Gopher",
    "email": "gordon@golang.org",
    "avatar_url": "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87"
  }
}`

// HookPullRequest is a sample pull_request webhook payload
const HookPullRequest = `{
  "action": "opened",
  "number": 1,
  "pull_request": {
    "html_url": "http://forgejo.golang.org/gordon/hello-world/pull/1",
    "state": "open",
    "title": "Update the README with new information",
    "body": "please merge",
    "user": {
      "id": 1,
      "username": "gordon",
      "login": "gordon",
      "full_name": "Gordon the Gopher",
      "email": "gordon@golang.org",
      "avatar_url": "http://forgejo.golang.org///1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87"
    },
    "base": {
      "label": "master",
      "ref": "master",
      "sha": "9353195a19e45482665306e466c832c46560532d"
    },
    "head": {
      "label": "feature/changes",
      "ref": "feature/changes",
      "sha": "0d1a26e67d8f5eaf1f6ba5c57fc3c7d91ac0fd1c"
    }
  },
  "repository": {
    "id": 35129377,
    "name": "hello-world",
    "full_name": "gordon/hello-world",
    "owner": {
      "id": 1,
      "username": "gordon",
      "login": "gordon",
      "full_name": "Gordon the Gopher",
      "email": "gordon@golang.org",
      "avatar_url": "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87"
    },
    "private": true,
    "html_url": "http://forgejo.golang.org/gordon/hello-world",
    "clone_url": "https://forgejo.golang.org/gordon/hello-world.git",
    "default_branch": "master"
  },
  "sender": {
      "id": 1,
      "login": "gordon",
      "username": "gordon",
      "full_name": "Gordon the Gopher",
      "email": "gordon@golang.org",
      "avatar_url": "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87"
    }
}`
