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
	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type oldRegistry007 struct {
	ID    int64  `json:"id"       xorm:"pk autoincr 'registry_id'"`
	Token string `json:"token"    xorm:"TEXT 'registry_token'"`
	Email string `json:"email"    xorm:"varchar(500) 'registry_email'"`
}

func (oldRegistry007) TableName() string {
	return "registry"
}

type oldPipeline007 struct {
	ID       int64  `json:"id"                      xorm:"pk autoincr 'pipeline_id'"`
	ConfigID int64  `json:"-"                       xorm:"pipeline_config_id"`
	Enqueued int64  `json:"enqueued_at"             xorm:"pipeline_enqueued"`
	CloneURL string `json:"clone_url"               xorm:"pipeline_clone_url"`
}

// TableName return database table name for xorm.
func (oldPipeline007) TableName() string {
	return "pipelines"
}

var cleanRegistryPipeline = xormigrate.Migration{
	ID: "clean-registry-pipeline",
	MigrateSession: func(sess *xorm.Session) (err error) {
		if err := sess.Sync(new(oldRegistry007), new(oldPipeline007)); err != nil {
			return err
		}

		if err := dropTableColumns(sess, "pipelines", "pipeline_clone_url", "pipeline_config_id", "pipeline_enqueued"); err != nil {
			return err
		}

		return dropTableColumns(sess, "registry", "registry_email", "registry_token")
	},
}
