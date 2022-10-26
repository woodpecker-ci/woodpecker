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
	"reflect"

	"github.com/rs/zerolog/log"
	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// APPEND NEW MIGRATIONS
// they are executed in order and if one fail woodpecker try to rollback and quit
var migrationTasks = []*task{
	&legacy2Xorm,
	&alterTableReposDropFallback,
	&alterTableReposDropAllowDeploysAllowTags,
	&fixPRSecretEventName,
	&alterTableReposDropCounter,
	&dropSenders,
	&alterTableLogUpdateColumnLogDataType,
	&alterTableSecretsAddUserCol,
	&lowercaseSecretNames,
	&renameBuildsToPipeline,
	&renameColumnsBuildsToPipeline,
	&renameTableProcsToSteps,
	&renameColumnsProcsToSteps,
	&renameColumnsJobsToSteps,
}

var allBeans = []interface{}{
	new(model.Agent),
	new(model.Pipeline),
	new(model.PipelineConfig),
	new(model.Config),
	new(model.File),
	new(model.Logs),
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
}

type migrations struct {
	Name string `xorm:"UNIQUE"`
}

type task struct {
	name     string
	required bool
	fn       func(sess *xorm.Session) error
}

// initNew create tables for new instance
func initNew(sess *xorm.Session) error {
	if err := syncAll(sess); err != nil {
		return err
	}

	// dummy run migrations
	for _, task := range migrationTasks {
		if _, err := sess.Insert(&migrations{task.name}); err != nil {
			return err
		}
	}

	return nil
}

func Migrate(e *xorm.Engine) error {
	if err := e.Sync2(new(migrations)); err != nil {
		return err
	}

	sess := e.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	// check if we have a fresh installation or need to check for migrations
	c, err := sess.Count(new(migrations))
	if err != nil {
		return err
	}

	if c == 0 {
		if err := initNew(sess); err != nil {
			return err
		}
		return sess.Commit()
	}

	if err := runTasks(sess, migrationTasks); err != nil {
		return err
	}

	if err := sess.Commit(); err != nil {
		return err
	}

	if err := e.ClearCache(allBeans...); err != nil {
		return err
	}

	return syncAll(e)
}

func runTasks(sess *xorm.Session, tasks []*task) error {
	// cache migrations in db
	migCache := make(map[string]bool)
	var migList []*migrations
	if err := sess.Find(&migList); err != nil {
		return err
	}
	for i := range migList {
		migCache[migList[i].Name] = true
	}

	for _, task := range tasks {
		if migCache[task.name] {
			log.Trace().Msgf("migration task '%s' already applied", task.name)
			continue
		}

		log.Trace().Msgf("start migration task '%s'", task.name)

		if task.fn != nil {
			if err := task.fn(sess); err != nil {
				if task.required {
					return err
				}
				log.Error().Err(err).Msgf("migration task '%s' failed but is not required", task.name)
				continue
			}
			log.Debug().Msgf("migration task '%s' done", task.name)
		} else {
			log.Trace().Msgf("skip migration task '%s'", task.name)
		}

		if _, err := sess.Insert(&migrations{task.name}); err != nil {
			return err
		}
		migCache[task.name] = true
	}
	return nil
}

type syncEngine interface {
	Sync2(beans ...interface{}) error
}

func syncAll(sess syncEngine) error {
	for _, bean := range allBeans {
		if err := sess.Sync2(bean); err != nil {
			return fmt.Errorf("sync2 error '%s': %v", reflect.TypeOf(bean), err)
		}
	}
	return nil
}
