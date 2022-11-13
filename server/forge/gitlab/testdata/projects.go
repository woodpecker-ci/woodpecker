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

package testdata

// sample repository list
var allProjectsPayload = []byte(`
[
	{
		"id": 4,
		"description": null,
		"default_branch": "master",
		"public": false,
		"visibility_level": 0,
		"ssh_url_to_repo": "git@example.com:diaspora/diaspora-client.git",
		"http_url_to_repo": "http://example.com/diaspora/diaspora-client.git",
		"web_url": "http://example.com/diaspora/diaspora-client",
		"owner": {
			"id": 3,
			"name": "Diaspora",
			"username": "some_user",
			"created_at": "2013-09-30T13:46:02Z"
		},
		"name": "Diaspora Client",
		"name_with_namespace": "Diaspora / Diaspora Client",
		"path": "diaspora-client",
		"path_with_namespace": "diaspora/diaspora-client",
		"issues_enabled": true,
		"merge_requests_enabled": true,
		"wiki_enabled": true,
		"snippets_enabled": false,
		"created_at": "2013-09-30T13:46:02Z",
		"last_activity_at": "2013-09-30T13:46:02Z",
		"namespace": {
			"created_at": "2013-09-30T13:46:02Z",
			"description": "",
			"id": 3,
			"name": "Diaspora",
			"owner_id": 1,
			"path": "diaspora",
			"updated_at": "2013-09-30T13:46:02Z"
		},
		"archived": false
	},
	{
		"id": 6,
		"description": null,
		"default_branch": "master",
		"public": false,
		"visibility_level": 0,
		"ssh_url_to_repo": "git@example.com:brightbox/puppet.git",
		"http_url_to_repo": "http://example.com/brightbox/puppet.git",
		"web_url": "http://example.com/brightbox/puppet",
		"owner": {
			"id": 1,
			"name": "Brightbox",
			"username": "test_user",
			"created_at": "2013-09-30T13:46:02Z"
		},
		"name": "Puppet",
		"name_with_namespace": "Brightbox / Puppet",
		"path": "puppet",
		"path_with_namespace": "brightbox/puppet",
		"issues_enabled": true,
		"merge_requests_enabled": true,
		"wiki_enabled": true,
		"snippets_enabled": false,
		"created_at": "2013-09-30T13:46:02Z",
		"last_activity_at": "2013-09-30T13:46:02Z",
		"namespace": {
			"created_at": "2013-09-30T13:46:02Z",
			"description": "",
			"id": 4,
			"name": "Brightbox",
			"owner_id": 1,
			"path": "brightbox",
			"updated_at": "2013-09-30T13:46:02Z"
		},
		"archived": true
	}
]
`)

var notArchivedProjectsPayload = []byte(`
[
	{
		"id": 4,
		"description": null,
		"default_branch": "master",
		"public": false,
		"visibility_level": 0,
		"ssh_url_to_repo": "git@example.com:diaspora/diaspora-client.git",
		"http_url_to_repo": "http://example.com/diaspora/diaspora-client.git",
		"web_url": "http://example.com/diaspora/diaspora-client",
		"owner": {
			"id": 3,
			"name": "Diaspora",
			"username": "some_user",
			"created_at": "2013-09-30T13:46:02Z"
		},
		"name": "Diaspora Client",
		"name_with_namespace": "Diaspora / Diaspora Client",
		"path": "diaspora-client",
		"path_with_namespace": "diaspora/diaspora-client",
		"issues_enabled": true,
		"merge_requests_enabled": true,
		"wiki_enabled": true,
		"snippets_enabled": false,
		"created_at": "2013-09-30T13:46:02Z",
		"last_activity_at": "2013-09-30T13:46:02Z",
		"namespace": {
			"created_at": "2013-09-30T13:46:02Z",
			"description": "",
			"id": 3,
			"name": "Diaspora",
			"owner_id": 1,
			"path": "diaspora",
			"updated_at": "2013-09-30T13:46:02Z"
		},
		"archived": false
	}
]
`)

var project4Payload = []byte(`
{
	"id": 4,
	"description": null,
	"default_branch": "master",
	"public": false,
	"visibility_level": 0,
	"ssh_url_to_repo": "git@example.com:diaspora/diaspora-client.git",
	"http_url_to_repo": "http://example.com/diaspora/diaspora-client.git",
	"web_url": "http://example.com/diaspora/diaspora-client",
	"owner": {
		"id": 3,
		"name": "Diaspora",
		"username": "some_user",
		"created_at": "2013-09-30T13:46:02Z"
	},
	"name": "Diaspora Client",
	"name_with_namespace": "Diaspora / Diaspora Client",
	"path": "diaspora-client",
	"path_with_namespace": "diaspora/diaspora-client",
	"issues_enabled": true,
	"merge_requests_enabled": true,
	"wiki_enabled": true,
	"snippets_enabled": false,
	"created_at": "2013-09-30T13:46:02Z",
	"last_activity_at": "2013-09-30T13:46:02Z",
	"namespace": {
		"created_at": "2013-09-30T13:46:02Z",
		"description": "",
		"id": 3,
		"name": "Diaspora",
		"owner_id": 1,
		"path": "diaspora",
		"updated_at": "2013-09-30T13:46:02Z"
	},
	"archived": false,
	"permissions": {
		"project_access": {
			"access_level": 10,
			"notification_level": 3
		},
		"group_access": {
			"access_level": 50,
			"notification_level": 3
		}
	}
}
`)

var project6Payload = []byte(`
{
	"id": 6,
	"description": null,
	"default_branch": "master",
	"public": false,
	"visibility_level": 0,
	"ssh_url_to_repo": "git@example.com:brightbox/puppet.git",
	"http_url_to_repo": "http://example.com/brightbox/puppet.git",
	"web_url": "http://example.com/brightbox/puppet",
	"owner": {
		"id": 1,
		"name": "Brightbox",
		"username": "test_user",
		"created_at": "2013-09-30T13:46:02Z"
	},
	"name": "Puppet",
	"name_with_namespace": "Brightbox / Puppet",
	"path": "puppet",
	"path_with_namespace": "brightbox/puppet",
	"issues_enabled": true,
	"merge_requests_enabled": true,
	"wiki_enabled": true,
	"snippets_enabled": false,
	"created_at": "2013-09-30T13:46:02Z",
	"last_activity_at": "2013-09-30T13:46:02Z",
	"namespace": {
		"created_at": "2013-09-30T13:46:02Z",
		"description": "",
		"id": 4,
		"name": "Brightbox",
		"owner_id": 1,
		"path": "brightbox",
		"updated_at": "2013-09-30T13:46:02Z"
	},
	"archived": false,
	"permissions": {
		"project_access": null,
		"group_access": null
	}
}
`)

var project4PayloadHook = []byte(`
{
	"id": 10717088,
	"url": "http://example.com/api/hook",
	"created_at": "2021-12-18T23:29:33.852Z",
	"push_events": true,
	"tag_push_events": true,
	"merge_requests_events": true,
	"repository_update_events": false,
	"enable_ssl_verification": true,
	"project_id": 4,
	"issues_events": false,
	"confidential_issues_events": false,
	"note_events": false,
	"confidential_note_events": null,
	"pipeline_events": false,
	"wiki_page_events": false,
	"deployment_events": true,
	"job_events": false,
	"releases_events": false,
	"push_events_branch_filter": null
}
`)

var project4PayloadHooks = []byte(`
[
  {
    "id": 10717088,
    "url": "http://example.com/api/hook",
    "created_at": "2021-12-18T23:29:33.852Z",
    "push_events": true,
    "tag_push_events": true,
    "merge_requests_events": true,
    "repository_update_events": false,
    "enable_ssl_verification": true,
    "project_id": 4,
    "issues_events": false,
    "confidential_issues_events": false,
    "note_events": false,
    "confidential_note_events": null,
    "pipeline_events": false,
    "wiki_page_events": false,
    "deployment_events": true,
    "job_events": false,
    "releases_events": false,
    "push_events_branch_filter": null
  }
]
`)
