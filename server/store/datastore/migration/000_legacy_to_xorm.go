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

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

var legacy2Xorm = task{
	name:     "xorm",
	required: true,
	fn: func(sess *xorm.Session) error {
		// make sure we have required migrations - else fail and point to last major version
		for _, mig := range []string{
			// users
			"create-table-users",
			"update-table-set-users-token-and-secret-length",
			// repos
			"create-table-repos",
			"alter-table-add-repo-visibility",
			"update-table-set-repo-visibility",
			"alter-table-add-repo-seq",
			"update-table-set-repo-seq",
			"update-table-set-repo-seq-default",
			"alter-table-add-repo-active",
			"update-table-set-repo-active",
			"alter-table-add-repo-fallback", // needed to drop col
			// builds
			"create-table-builds",
			"create-index-builds-repo",
			"create-index-builds-author",
			// procs
			"create-table-procs",
			"create-index-procs-build",
			// files
			"create-table-files",
			"create-index-files-builds",
			"create-index-files-procs",
			"alter-table-add-file-pid",
			"alter-table-add-file-meta-passed",
			"alter-table-add-file-meta-failed",
			"alter-table-add-file-meta-skipped",
			"alter-table-update-file-meta",
			// secrets
			"create-table-secrets",
			"create-index-secrets-repo",
			// registry
			"create-table-registry",
			"create-index-registry-repo",
			// senders
			"create-table-senders",
			"create-index-sender-repos",
			// perms
			"create-table-perms",
			"create-index-perms-repo",
			"create-index-perms-user",
			// build_config
			"create-table-build-config",
			"populate-build-config",
		} {
			exist, err := sess.Exist(&migrations{mig})
			if err != nil {
				return fmt.Errorf("test migration existence: %v", err)
			}
			if !exist {
				log.Error().Msgf("migration step '%s' missing, please upgrade to last stable v0.14.x version first", mig)
				return fmt.Errorf("legacy migration step missing")
			}
		}

		{ // recreate build_config
			type BuildConfig struct {
				ConfigID int64 `xorm:"NOT NULL 'config_id'"` // xorm.Sync2() do not use index info of sess -> so it try to create it twice
				BuildID  int64 `xorm:"NOT NULL 'build_id'"`
			}
			if err := renameTable(sess, "build_config", "old_build_config"); err != nil {
				return err
			}
			if err := sess.Sync2(new(BuildConfig)); err != nil {
				return err
			}
			if _, err := sess.Exec("INSERT INTO build_config (config_id, build_id) SELECT config_id,build_id FROM old_build_config;"); err != nil {
				return fmt.Errorf("unable to set copy data into temp table %s. Error: %v", "old_build_config", err)
			}
			if err := sess.DropTable("old_build_config"); err != nil {
				return fmt.Errorf("could not drop table '%s': %v", "old_build_config", err)
			}
		}

		dialect := sess.Engine().Dialect().URI().DBType
		switch dialect {
		case schemas.MYSQL:
			for _, exec := range []string{
				"DROP INDEX IF EXISTS build_number ON builds;",
				"DROP INDEX IF EXISTS ix_build_repo ON builds;",
				"DROP INDEX IF EXISTS ix_build_author ON builds;",
				"DROP INDEX IF EXISTS proc_build_ix ON procs;",
				"DROP INDEX IF EXISTS file_build_ix ON files;",
				"DROP INDEX IF EXISTS file_proc_ix  ON files;",
				"DROP INDEX IF EXISTS ix_secrets_repo  ON secrets;",
				"DROP INDEX IF EXISTS ix_registry_repo ON registry;",
				"DROP INDEX IF EXISTS sender_repo_ix ON senders;",
				"DROP INDEX IF EXISTS ix_perms_repo ON perms;",
				"DROP INDEX IF EXISTS ix_perms_user ON perms;",
			} {
				if _, err := sess.Exec(exec); err != nil {
					return fmt.Errorf("exec: '%s' failed: %v", exec, err)
				}
			}
		case schemas.SQLITE, schemas.POSTGRES:
			for _, exec := range []string{
				"DROP INDEX IF EXISTS ix_build_status_running;",
				"DROP INDEX IF EXISTS ix_build_repo;",
				"DROP INDEX IF EXISTS ix_build_author;",
				"DROP INDEX IF EXISTS proc_build_ix;",
				"DROP INDEX IF EXISTS file_build_ix;",
				"DROP INDEX IF EXISTS file_proc_ix;",
				"DROP INDEX IF EXISTS ix_secrets_repo;",
				"DROP INDEX IF EXISTS ix_registry_repo;",
				"DROP INDEX IF EXISTS sender_repo_ix;",
				"DROP INDEX IF EXISTS ix_perms_repo;",
				"DROP INDEX IF EXISTS ix_perms_user;",
			} {
				if _, err := sess.Exec(exec); err != nil {
					return fmt.Errorf("exec: '%s' failed: %v", exec, err)
				}
			}
		default:
			return fmt.Errorf("dialect '%s' not supported", dialect)
		}

		return nil
	},
}
