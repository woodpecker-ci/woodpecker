// Copyright 2026 Woodpecker Authors
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

package model

import "time"

// Session holds a web session.
type Session struct {
	ID      int64  `       xorm:"pk autoincr 'id'"`
	UserID  int64  `  xorm:"NOT NULL 'user_id'"`
	Created int64  `xorm:"'created' NOT NULL DEFAULT 0 created"`
	Token   string `xorm:"token"`
}

func (Session) TableName() string {
	return "sessions"
}

func (s Session) Expired(expires time.Duration) bool {
	return s.Created > time.Now().Add(-expires).Unix()
}
