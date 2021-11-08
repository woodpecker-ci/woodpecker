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

	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

// APPEND NEW MIGRATIONS
var migrationTasks = []task{
	{
		name: "xorm",
		fn:   nil,
	},
}

type migrations struct {
	Name string `xorm:"UNIQUE"`
}

type task struct {
	name string
	fn   func(sess *xorm.Session) error
}

// initNew create tables for new instance
func initNew(e *xorm.Engine) error {
	sess := e.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	if err := syncAll(sess); err != nil {
		return err
	}

	if _, err := sess.Insert(&migrations{"xorm"}); err != nil {
		return err
	}

	return sess.Commit()
}

func Migrate(e *xorm.Engine) error {
	if err := e.Sync2(new(migrations)); err != nil {
		return err
	}

	// check if we have a fresh installation or need to check for migrations
	c, err := e.Count(new(migrations))
	if err != nil {
		return err
	}

	if c == 0 {
		return initNew(e)
	}

	// handle old instance
	noLegacy, err := e.Exist(&migrations{"xorm"})
	if err != nil {
		return err
	}
	if !noLegacy {
		if err := legacyMigrations(e); err != nil {
			return err
		}
	}

	if err := runTasks(e, migrationTasks); err != nil {
		return err
	}

	sess := e.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}
	if err := syncAll(sess); err != nil {
		return err
	}
	return sess.Commit()
}

func runTasks(e *xorm.Engine, tasks []task) error {
	sess := e.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	for _, task := range tasks {
		exist, err := sess.Exist(&migrations{task.name})
		if err != nil {
			return err
		}
		if exist {
			continue
		}

		if task.fn != nil {
			if err := task.fn(sess); err != nil {
				return err
			}
		}

		if _, err := sess.Insert(&migrations{task.name}); err != nil {
			return err
		}
	}

	return sess.Commit()
}

func syncAll(sess *xorm.Session) error {
	for _, bean := range []interface{}{
		new(model.Agent),
		new(model.Build),
		new(model.BuildConfig),
		new(model.Config),
		new(model.File),
		new(model.Logs),
		new(model.Perm),
		new(model.Proc),
		new(model.Registry),
		new(model.Repo),
		new(model.Secret),
		new(model.Sender),
		new(model.Task),
		new(model.User),
	} {
		if err := sess.Sync2(bean); err != nil {
			return fmt.Errorf("sync2 error '%s': %v", reflect.TypeOf(bean), err)
		}
	}
	return nil
}
