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
)

var addInfraFailureRetry = xormigrate.Migration{
	ID: "add-infra-failure-retry",
	MigrateSession: func(sess *xorm.Session) error {
		type steps struct {
			ID int64 `xorm:"pk autoincr 'id'"`

			// marks a step terminated by an infrastructure event
			InfraFailure bool `xorm:"NOT NULL DEFAULT false 'infra_failure'"`
		}

		type pipelines struct {
			ID int64 `xorm:"pk autoincr 'id'"`

			// number of automatic infrastructure-failure retries that led to
			// this pipeline; bounds WOODPECKER_INFRA_RETRY_MAX_ATTEMPTS
			InfraRetryCount int64 `xorm:"NOT NULL DEFAULT 0 'infra_retry_count'"`

			// claimed (atomically) by the single Done that triggers this
			// pipeline's automatic infra retry, so concurrent Done calls
			// cannot each spawn a duplicate restart
			InfraRetried bool `xorm:"NOT NULL DEFAULT false 'infra_retried'"`
		}

		if err := sess.Sync(new(steps)); err != nil {
			return err
		}

		return sess.Sync(new(pipelines))
	},
}
