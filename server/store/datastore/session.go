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

package datastore

import (
	"time"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

// GetSession gets session by token
func (s storage) GetSession(token string) (*model.Session, error) {
	session := new(model.Session)
	return session, wrapGet(s.engine.Where("token = ?", token).Get(session))
}

// CreateSession creates a new session
func (s storage) CreateSession(session *model.Session) error {
	_, err := s.engine.Insert(session)
	return err
}

// DeleteSession deletes a session by token
func (s storage) DeleteSession(token string) error {
	return wrapDelete(s.engine.Where("token = ?", token).Delete(new(model.Session)))
}

// DeleteExpiredSessions deletes all expired sessions
func (s storage) DeleteExpiredSessions() error {
	return wrapDelete(s.engine.Where("created < ?", time.Now().Add(-server.Config.Server.SessionExpires).Unix()).Delete(new(model.Session)))
}
