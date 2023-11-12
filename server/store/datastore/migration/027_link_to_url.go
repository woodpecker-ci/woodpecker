// Copyright 2023 Woodpecker Authors
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
	"xorm.io/xorm"
)

var renameLinkToURL = task{
	name:     "rename-link-to-url",
	required: true,
	fn: func(sess *xorm.Session) (err error) {
		if err := renameColumn(sess, "pipelines", "pipeline_link", "pipeline_url"); err != nil {
			return err
		}

		return renameColumn(sess, "repos", "repo_link", "repo_url")
	},
}
