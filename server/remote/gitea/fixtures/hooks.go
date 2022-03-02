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

// HookPush is a sample Gitea push hook
const HookPush = `
{
  "ref": "refs/heads/master",
  "before": "4b2626259b5a97b6b4eab5e6cca66adb986b672b",
  "after": "ef98532add3b2feb7a137426bba1248724367df5",
  "compare_url": "http://gitea.golang.org/gordon/hello-world/compare/4b2626259b5a97b6b4eab5e6cca66adb986b672b...ef98532add3b2feb7a137426bba1248724367df5",
  "commits": [
    {
      "id": "ef98532add3b2feb7a137426bba1248724367df5",
      "message": "bump\n",
      "url": "http://gitea.golang.org/gordon/hello-world/commit/ef98532add3b2feb7a137426bba1248724367df5",
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
    "html_url": "http://gitea.golang.org/gordon/hello-world",
    "ssh_url": "git@gitea.golang.org:gordon/hello-world.git",
    "clone_url": "http://gitea.golang.org/gordon/hello-world.git",
    "description": "",
    "website": "",
    "watchers": 1,
    "owner": {
      "name": "gordon",
      "email": "gordon@golang.org",
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
    "avatar_url": "http://gitea.golang.org///1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87"
  }
}
`

// HookPushTag is a sample Gitea tag hook
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
      "full_name": "Gordon the Gopher",
      "email": "gordon@golang.org",
      "avatar_url": "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87"
    },
    "name": "hello-world",
    "full_name": "gordon/hello-world",
    "description": "",
    "private": true,
    "fork": false,
    "html_url": "http://gitea.golang.org/gordon/hello-world",
    "ssh_url": "git@gitea.golang.org:gordon/hello-world.git",
    "clone_url": "http://gitea.golang.org/gordon/hello-world.git",
    "default_branch": "master",
    "created_at": "2015-10-22T19:32:44Z",
    "updated_at": "2016-11-24T13:37:16Z"
  },
  "sender": {
    "id": 1,
    "username": "gordon",
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
    "html_url": "http://gitea.golang.org/gordon/hello-world/pull/1",
    "state": "open",
    "title": "Update the README with new information",
    "body": "please merge",
    "user": {
      "id": 1,
      "username": "gordon",
      "full_name": "Gordon the Gopher",
      "email": "gordon@golang.org",
      "avatar_url": "http://gitea.golang.org///1.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87"
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
      "full_name": "Gordon the Gopher",
      "email": "gordon@golang.org",
      "avatar_url": "https://secure.gravatar.com/avatar/8c58a0be77ee441bb8f8595b7f1b4e87"
    },
    "private": true,
    "html_url": "http://gitea.golang.org/gordon/hello-world",
    "clone_url": "https://gitea.golang.org/gordon/hello-world.git",
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
