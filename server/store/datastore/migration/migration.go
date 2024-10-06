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
	"context"
	"fmt"
	"reflect"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

// APPEND NEW MIGRATIONS
// They are executed in order and if one fails Xormigrate will try to rollback that specific one and quits.
var migrationTasks = []*xormigrate.Migration{
	&legacyToXormigrate,
	&addOrgID,
	&alterTableTasksUpdateColumnTaskDataType,
	&alterTableConfigUpdateColumnConfigDataType,
	&removePluginOnlyOptionFromSecretsTable,
	&convertToNewPipelineErrorFormat,
	&renameLinkToURL,
	&cleanRegistryPipeline,
	&setForgeID,
	&unifyColumnsTables,
	&alterTableRegistriesFixRequiredFields,
	&cronWithoutSec,
	&renameStartEndTime,
	&fixV31Registries,
	&removeOldMigrationsOfV1,
	&addOrgAgents,
	&addCustomLabelsToAgent,
}

var allBeans = []any{
	new(model.Agent),
	new(model.Pipeline),
	new(model.PipelineConfig),
	new(model.Config),
	new(model.LogEntry),
	new(model.Perm),
	new(model.Step),
	new(model.Registry),
	new(model.Repo),
	new(model.Secret),
	new(model.Task),
	new(model.User),
	new(model.ServerConfig),
	new(model.Cron),
	new(model.Redirection),
	new(model.Forge),
	new(model.Workflow),
	new(model.Org),
}

// TODO: make xormigrate context aware
func Migrate(_ context.Context, e *xorm.Engine, allowLong bool) error {
	e.SetDisableGlobalCache(true)

	m := xormigrate.New(e, migrationTasks)
	m.AllowLong(allowLong)
	oldCount, err := e.Table("migrations").Count()
	if oldCount < 1 || err != nil {
		// allow new schema initialization if old migrations table is empty or it does not exist (err != nil)
		// schema initialization will always run if we call `InitSchema`
		m.InitSchema(func(_ *xorm.Engine) error {
			// do nothing on schema init, models are synced in any case below
			return nil
		})
	}

	m.SetLogger(&xormigrateLogger{})

	if err := m.Migrate(); err != nil {
		return err
	}

	e.SetDisableGlobalCache(false)

	if err := syncAll(e); err != nil {
		return fmt.Errorf("msg: %w", err)
	}

	return nil
}

func syncAll(sess *xorm.Engine) error {
	for _, bean := range allBeans {
		if err := sess.Sync(bean); err != nil {
			return fmt.Errorf("sync error '%s': %w", reflect.TypeOf(bean), err)
		}
	}
	return nil
}
