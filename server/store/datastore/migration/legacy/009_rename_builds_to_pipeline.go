// Copyright 2022 Woodpecker Authors
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

package legacy

import (
	"xorm.io/xorm"
)

var renameBuildsToPipeline = task{
	name:     "rename-builds-to-pipeline",
	required: true,
	fn: func(sess *xorm.Session) error {
		err := renameTable(sess, "builds", "pipelines")
		if err != nil {
			return err
		}
		err = renameTable(sess, "build_config", "pipeline_config")
		if err != nil {
			return err
		}
		return nil
	},
}
