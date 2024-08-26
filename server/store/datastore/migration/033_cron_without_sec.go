// Copyright 2024 Woodpecker Authors
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
	"fmt"
	"strings"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

var cronWithoutSec = xormigrate.Migration{
	ID: "cron-without-sec",
	MigrateSession: func(sess *xorm.Session) error {
		if err := sess.Sync(new(model.Cron)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		var crons []*model.Cron
		if err := sess.Find(&crons); err != nil {
			return err
		}

		for _, c := range crons {
			if strings.HasPrefix(strings.TrimSpace(c.Schedule), "@") {
				// something like "@daily"
				continue
			}

			if _, err := sess.Update(&model.Cron{
				Schedule: strings.SplitN(strings.TrimSpace(c.Schedule), " ", 2)[1],
			}, c); err != nil {
				return err
			}
		}

		return nil
	},
}
