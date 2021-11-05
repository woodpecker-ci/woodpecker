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

package migration

import "xorm.io/xorm"

var legacyMigrations = []task{
	{
		name: "create-table-users",
		fn:   nil,
	},
	{
		name: "create-table-repos",
		fn:   nil,
	},
	{
		name: "create-table-builds",
		fn:   nil,
	},
	{
		name: "create-index-builds-repo",
		fn:   nil,
	},
	{
		name: "create-index-builds-author",
		fn:   nil,
	},
	{
		name: "create-table-procs",
		fn:   nil,
	},
	{
		name: "create-index-procs-build",
		fn:   nil,
	},
	{
		name: "create-table-logs",
		fn:   nil,
	},
	{
		name: "create-table-files",
		fn:   nil,
	},
	{
		name: "create-index-files-builds",
		fn:   nil,
	},
	{
		name: "create-index-files-procs",
		fn:   nil,
	},
	{
		name: "create-table-secrets",
		fn:   nil,
	},
	{
		name: "create-index-secrets-repo",
		fn:   nil,
	},
	{
		name: "create-table-registry",
		fn:   nil,
	},
	{
		name: "create-index-registry-repo",
		fn:   nil,
	},
	{
		name: "create-table-config",
		fn:   nil,
	},
	{
		name: "create-table-tasks",
		fn:   nil,
	},
	{
		name: "create-table-agents",
		fn: func(sess *xorm.Session) error {
			return sess.Sync2(new(legacyAgent))
		},
	},
	{
		name: "create-table-senders",
		fn:   nil,
	},
	{
		name: "create-index-sender-repos",
		fn:   nil,
	},
	{
		name: "alter-table-add-repo-visibility",
		fn:   nil,
	},
	{
		name: "update-table-set-repo-visibility",
		fn:   nil,
	},
	{
		name: "alter-table-add-repo-seq",
		fn:   nil,
	},
	{
		name: "update-table-set-repo-seq",
		fn:   nil,
	},
	{
		name: "update-table-set-repo-seq-default",
		fn:   nil,
	},
	{
		name: "alter-table-add-repo-active",
		fn:   nil,
	},
	{
		name: "update-table-set-repo-active",
		fn:   nil,
	},
	{
		name: "alter-table-add-user-synced",
		fn:   nil,
	},
	{
		name: "update-table-set-user-synced",
		fn:   nil,
	},
	{
		name: "create-table-perms",
		fn:   nil,
	},
	{
		name: "create-index-perms-repo",
		fn:   nil,
	},
	{
		name: "create-index-perms-user",
		fn:   nil,
	},
	{
		name: "alter-table-add-file-pid",
		fn:   nil,
	},
	{
		name: "alter-table-add-file-meta-passed",
		fn:   nil,
	},
	{
		name: "alter-table-add-file-meta-failed",
		fn:   nil,
	},
	{
		name: "alter-table-add-file-meta-skipped",
		fn:   nil,
	},
	{
		name: "alter-table-update-file-meta",
		fn:   nil,
	},
	{
		name: "create-table-build-config",
		fn:   nil,
	},
	{
		name: "alter-table-add-config-name",
		fn:   nil,
	},
	{
		name: "update-table-set-config-name",
		fn:   nil,
	},
	{
		name: "populate-build-config",
		fn:   nil,
	},
	{
		name: "alter-table-add-task-dependencies",
		fn:   nil,
	},
	{
		name: "alter-table-add-task-run-on",
		fn:   nil,
	},
	{
		name: "alter-table-add-repo-fallback",
		fn:   nil,
	},
	{
		name: "update-table-set-repo-fallback",
		fn:   nil,
	},
	{
		name: "update-table-set-repo-fallback-again",
		fn:   nil,
	},
	{
		name: "add-builds-changed_files-column",
		fn:   nil,
	},
	{
		name: "update-builds-set-changed_files",
		fn:   nil,
	},
	{
		name: "alter-table-drop-repo-fallback",
		fn:   nil,
	},
	{
		name: "drop-allow-push-tags-deploys-columns",
		fn:   nil,
	},
	{
		name: "update-table-set-users-token-and-secret-length",
		fn:   nil,
	},
}

type legacyAgent struct {
	ID       int64  `xorm:"pk autoincr 'agent_id'"`
	Addr     string `xorm:"UNIQUE VARCHAR(250) 'agent_addr'"`
	Platform string `xorm:"VARCHAR(500) 'agent_platform'"`
	Capacity int64  `xorm:"agent_capacity"`
	Created  int64  `xorm:"created 'agent_created'"`
	Updated  int64  `xorm:"updated 'agent_updated'"`
}
