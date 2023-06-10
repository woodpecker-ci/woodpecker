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
	"errors"
	"fmt"
	"reflect"

	"github.com/rs/zerolog/log"
	"xorm.io/xorm"
)

var ErrDestHasExistingData = errors.New("while try to convert the database, we detect existing data at the destination")

func Copy(src, dest *xorm.Engine) error {
	// first check if the new database already has existing data
	for _, bean := range allBeans {
		exist, err := dest.IsTableExist(bean)
		if err != nil {
			return err
		} else if exist {
			return fmt.Errorf("%w: table %s", ErrDestHasExistingData, dest.TableName(bean))
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
	for _, bean := range allBeans {
		log.Info().Msgf("Start copy %s table", dest.TableName(bean))
		// TODO: add showBeAliveSign from #1846

		page := 0
		perPage := 10

		for {
			// TODO: use waitGroup and chanel to stream items through

			// create list out of type from bean
			beanType := reflect.TypeOf(bean)
			beanSliceType := reflect.New(reflect.SliceOf(beanType))
			items := beanSliceType.Interface()

			// FIXIT: ""needs a pointer to a slice or a map""

			// read
			if err := src.Limit(perPage, page*perPage).Find(&items); err != nil {
				return err
			}
			// write
			if _, err := dest.NoAutoTime().AllCols().InsertMulti(items); err != nil {
				return err
			}

			if reflect.ValueOf(items).Len() < perPage {
				break
			}
			page++
		}
	}

	return nil
}
