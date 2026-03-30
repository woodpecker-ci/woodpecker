// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package migration

import (
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

var addCronField = xormigrate.Migration{
	ID: "add-cron-field",
	MigrateSession: func(sess *xorm.Session) error {
		// perPage027 set the size of the slice to read per page.
		perPage027 := 100

		type pipelines struct {
			ID int64 `xorm:"pk autoincr 'id'"`

			Sender string `xorm:"sender"` // uses name of cron for cron pipelines

			// new fields
			Cron string `xorm:"cron"`
		}

		if err := sess.Sync(new(pipelines)); err != nil {
			return err
		}

		page := 0
		oldPipelines := make([]*pipelines, 0, perPage027)

		for {
			oldPipelines = oldPipelines[:0]

			err := sess.Limit(perPage027, page*perPage027).Where("event = ?", model.EventCron).Cols("id", "sender").Find(&oldPipelines)
			if err != nil {
				return err
			}

			for _, oldPipeline := range oldPipelines {
				newPipeline := pipelines{
					ID:   oldPipeline.ID,
					Cron: oldPipeline.Sender,
				}

				if _, err := sess.ID(oldPipeline.ID).Cols("cron").Update(newPipeline); err != nil {
					return err
				}
			}

			if len(oldPipelines) < perPage027 {
				break
			}

			page++
		}

		return nil
	},
}
