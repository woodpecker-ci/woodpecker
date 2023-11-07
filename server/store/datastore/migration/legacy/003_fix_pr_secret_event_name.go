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

package legacy

import (
	"xorm.io/xorm"

	"go.woodpecker-ci.org/woodpecker/server/model"
)

var fixPRSecretEventName = task{
	name: "fix-pr-secret-event-name",
	fn: func(sess *xorm.Session) error {
		const batchSize = 100
		for start := 0; ; start += batchSize {
			secrets := make([]*model.Secret, 0, batchSize)
			if err := sess.Limit(batchSize, start).Table("secrets").Cols("secret_id", "secret_events").Where("secret_events LIKE '%pull-request%'").Find(&secrets); err != nil {
				return err
			}

			if len(secrets) == 0 {
				break
			}

			for _, secret := range secrets {
				for i, event := range secret.Events {
					if event == "pull-request" {
						secret.Events[i] = "pull_request"
					}
				}
				if _, err := sess.ID(secret.ID).Cols("secret_events").Update(secret); err != nil {
					return err
				}
			}
		}
		return nil
	},
}
