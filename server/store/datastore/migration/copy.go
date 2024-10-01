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
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

const perPage = 1000

func Copy(ctx context.Context, src, dest *xorm.Engine) error {
	// first check if the new database already has existing data
	for _, bean := range AllBeans {
		exist, err := dest.IsTableExist(bean)
		if err != nil {
			return err
		} else if exist {
			return fmt.Errorf("existing table '%s' in import destination detected", dest.TableName(bean))
		}
	}

	// next we make sure the all required migrations are executed
	if err := Migrate(ctx, src, true); err != nil {
		return fmt.Errorf("migrate source database failed: %w", err)
	}

	// init schema in destination
	if err := initSchemaOnly(dest); err != nil {
		return err
	}

	// copy data
	return CopyData(ctx, src, dest)
}

func copyBean[T any](ctx context.Context, src, dest *xorm.Engine) error {
	tableName := dest.TableName(new(T))
	log.Info().Msgf("Start copy %s table", tableName)
	aliveMsgCancel := showBeAliveSign(ctx, tableName)
	defer aliveMsgCancel(nil)

	page := 0
	items := make([]*T, 0, perPage)

	for {
		// for each loop we just check if the context got canceled
		if err := ctx.Err(); err != nil {
			return err
		}

		// clean item list
		items = items[:0]
		log.Trace().Msgf("copy table '%s' page %d", tableName, page)

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

	// for postgress we have to manually update the sequence for autoincrements
	if dest.Dialect().URI().DBType == schemas.POSTGRES {
		t, err := dest.TableInfo(new(T))
		if err != nil {
			return err
		}
		if t.AutoIncrement != "" {
			max := 0
			if _, err := dest.SQL(fmt.Sprintf("SELECT MAX(%s) FROM %s;", t.AutoIncrement, tableName)).Get(&max); err != nil {
				return fmt.Errorf("could not get max value to calc auto increments max for postgress: %w", err)
			}

			log.Debug().Msgf("for '%s' found auto increment '%s' with current max at '%d'", tableName, t.AutoIncrement, max)

			if max > 0 {
				sql := fmt.Sprintf("SELECT pg_catalog.setval('%s_%s_seq', %d, true)", tableName, t.AutoIncrement, max)
				if _, err := dest.Exec(sql); err != nil {
					log.Debug().Err(err).Msgf("could not exec: '%s'", sql)
					return fmt.Errorf("could not update auto increments for postgress: %w", err)
				}
			}
		}
	}

	aliveMsgCancel(nil)
	return nil
}

var showBeAliveSignDelay = time.Second * 20

func showBeAliveSign(ctx context.Context, taskName string) context.CancelCauseFunc {
	ctx, cancel := context.WithCancelCause(ctx)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(showBeAliveSignDelay):
				log.Info().Msgf("Migration '%s' is still running, please be patient", taskName)
			}
		}
	}()
	return cancel
}
