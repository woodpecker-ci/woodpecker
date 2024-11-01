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

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
)

type repoV035 struct {
	ID        int64                      `xorm:"pk autoincr 'id'"`
	IsTrusted bool                       `xorm:"'trusted'"`
	Trusted   model.TrustedConfiguration `xorm:"json 'trusted_conf'"`
}

func (repoV035) TableName() string {
	return "repos"
}

var splitTrusted = xormigrate.Migration{
	ID: "split-trusted",
	MigrateSession: func(sess *xorm.Session) error {
		if err := sess.Sync(new(repoV035)); err != nil {
			return fmt.Errorf("sync new models failed: %w", err)
		}

		if _, err := sess.Where("trusted = ?", false).Cols("trusted_conf").Update(&repoV035{
			Trusted: model.TrustedConfiguration{
				Network:  false,
				Security: false,
				Volumes:  false,
			},
		}); err != nil {
			return err
		}

		if _, err := sess.Where("trusted = ?", true).Cols("trusted_conf").Update(&repoV035{
			Trusted: model.TrustedConfiguration{
				Network:  true,
				Security: true,
				Volumes:  true,
			},
		}); err != nil {
			return err
		}

		if err := dropTableColumns(sess, "repos", "trusted"); err != nil {
			return err
		}

		if err := sess.Commit(); err != nil {
			return err
		}

		return renameColumn(sess, "repos", "trusted_conf", "trusted")
	},
}
