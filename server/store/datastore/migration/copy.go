// Copyright 2023 Woodpecker Authors
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

	"github.com/woodpecker-ci/woodpecker/server/model"
)

func Copy(src, dest *xorm.Engine) error {
	// first check if the new database already has existing data
	for _, bean := range allBeans {
		exist, err := dest.IsTableExist(bean)
		if err != nil {
			return err
		} else if exist {
			return fmt.Errorf("existing table '%s' in import destination detected", dest.TableName(bean))
		}
	}

	// next we make sure the all required migrations are executed
	if err := Migrate(src); err != nil {
		return fmt.Errorf("migrate source database failed: %w", err)
	}

	// init schema in destination
	if err := dest.Sync(new(migrations)); err != nil {
		return err
	}
	if err := initNew(dest); err != nil {
		return fmt.Errorf("init schema at destination failed: %w", err)
	}

	// copy data
	// IMPORTANT: if you add something here, also add it to migration.go allBeans slice
	{ // TODO: find a way to use reflection to be able to use allBeans
		if err := copyBean[model.Agent](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.Pipeline](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.PipelineConfig](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.Config](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.LogEntry](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.Perm](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.Step](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.Registry](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.Repo](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.Secret](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.Task](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.User](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.ServerConfig](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.Cron](src, dest); err != nil {
			return err
		}
		if err := copyBean[model.Redirection](src, dest); err != nil {
			return err
		}
	}

	return nil
}

func copyBean[T any](src, dest *xorm.Engine) error {
	tableName := dest.TableName(new(T))
	log.Info().Msgf("Start copy %s table", tableName)
	aliveMsgCancel := showBeAliveSign(tableName)
	defer aliveMsgCancel(nil)

	page := 0
	perPage := 100
	items := make([]*T, 0, perPage)

	for {
		// TODO: use waitGroup and chanel to stream items through
		// clean item list
		items = items[:0]

		// read
		if err := src.Limit(perPage, page*perPage).Find(&items); err != nil {
			return fmt.Errorf("read data of table '%s' page %d failed: %w", tableName, page, err)
		}

		if len(items) == 0 {
			break
		}

		// write
		if _, err := dest.NoAutoTime().AllCols().InsertMulti(items); err != nil {
			return fmt.Errorf("write data of table '%s' page %d failed: %w", tableName, page, err)
		}

		if len(items) < perPage {
			break
		}
		page++
	}

	aliveMsgCancel(nil)
	return nil
}
