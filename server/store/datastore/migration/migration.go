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
	"xorm.io/xorm"

	"github.com/woodpecker-ci/woodpecker/server/model"
)

type migrations struct {
	Name string
}

func Migrate(e *xorm.Engine) error {
	if err := e.Sync2(new(migrations)); err != nil {
		return err
	}

	// TODO: handle old instance

	// create tables for new instance
	if err := e.Sync2(
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
	); err != nil {
		return err
	}
	return nil
}
