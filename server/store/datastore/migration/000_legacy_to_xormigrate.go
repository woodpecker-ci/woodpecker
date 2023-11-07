package migration

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/server/store/datastore/migration/legacy"
)

type v000Migrations struct {
	Name string `xorm:"UNIQUE"`
}

func (m *v000Migrations) TableName() string {
	return "migrations"
}

var legacyToXormigrate = xormigrate.Migration{
	ID: "legacy-to-xormigrate",
	Migrate: func(engine *xorm.Engine) error {
		legacyMigrations := []string{
			"xorm",
			"alter-table-drop-repo-fallback",
			"drop-allow-push-tags-deploys-columns",
			"fix-pr-secret-event-name",
			"alter-table-drop-counter",
			"drop-senders",
			"alter-table-logs-update-type-of-data",
			"alter-table-add-secrets-user-id",
			"recreate-agents-table",
			"lowercase-secret-names",
			"rename-builds-to-pipeline",
			"rename-columns-builds-to-pipeline",
			"rename-procs-to-steps",
			"rename-remote-to-forge",
			"rename-forge-id-to-forge-remote-id",
			"remove-active-from-users",
			"remove-inactive-repos",
			"drop-files",
			"remove-machine-col",
			"parent-steps-to-workflows",
			"drop-old-col",
			"init-log_entries",
			//"migrate-logs-to-log_entries", -> not required so we just skip it
			"add-orgs",
			"add-org-id",
			"alter-table-tasks-update-type-of-task-data",
			"alter-table-config-update-type-of-config-data",
			"remove-plugin-only-option-from-secrets-table",
			"convert-to-new-pipeline-error-format",
		}

		// TODO remove in 3.x and move to MigrateSession
		if err := legacy.Migrate(engine); err != nil {
			return err
		}

		for _, mig := range legacyMigrations {
			exist, err := engine.Exist(&v000Migrations{mig})
			if err != nil {
				return fmt.Errorf("test migration existence: %w", err)
			}
			if !exist {
				log.Error().Msgf("migration step '%s' missing, please upgrade to last stable v2.x version first", mig)
				return fmt.Errorf("legacy migration step missing")
			}
		}

		return engine.DropTables("migrations")
	},
}
