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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userV029 struct {
	ID      int64  `xorm:"pk autoincr 'id'"`
	ForgeID int64  `xorm:"forge_id"`
	Login   string `xorm:"'login'"`
	Hash    string `xorm:"'hash'"`
}

func (userV029) TableName() string { return "users" }

type orgV029 struct {
	ID      int64  `xorm:"pk autoincr 'id'"`
	ForgeID int64  `xorm:"forge_id"`
	Name    string `xorm:"'name'"`
	IsUser  bool   `xorm:"is_user"`
}

func (orgV029) TableName() string { return "orgs" }

func TestReplaceZeroForgeIDsInUsers(t *testing.T) {
	engine, closeDB := testDB(t, true)
	defer closeDB()

	require.NoError(t, engine.Sync(new(userV029), new(orgV029)))

	// user + personal org left with forge_id=0 by legacy CLI/web-ui provisioning
	_, err := engine.Insert(&userV029{ForgeID: 0, Login: "broken", Hash: "h1"})
	require.NoError(t, err)
	_, err = engine.Insert(&orgV029{ForgeID: 0, Name: "broken", IsUser: true})
	require.NoError(t, err)
	// a healthy user/org already on the real forge must stay untouched
	_, err = engine.Insert(&userV029{ForgeID: 1, Login: "healthy", Hash: "h2"})
	require.NoError(t, err)
	_, err = engine.Insert(&orgV029{ForgeID: 2, Name: "other-forge", IsUser: false})
	require.NoError(t, err)

	sess := engine.NewSession()
	defer sess.Close()
	require.NoError(t, replaceZeroForgeIDsInUsers.MigrateSession(sess))
	require.NoError(t, sess.Commit())

	zeroUsers, err := engine.Where("forge_id = 0").Count(new(userV029))
	require.NoError(t, err)
	assert.EqualValues(t, 0, zeroUsers, "no user should be left on forge_id=0")

	zeroOrgs, err := engine.Where("forge_id = 0").Count(new(orgV029))
	require.NoError(t, err)
	assert.EqualValues(t, 0, zeroOrgs, "no org should be left on forge_id=0")

	// the healed user is now resolvable on the default forge
	healed := new(userV029)
	found, err := engine.Where("login = ?", "broken").Get(healed)
	require.NoError(t, err)
	require.True(t, found)
	assert.EqualValues(t, 1, healed.ForgeID)

	// unrelated forge ids are untouched
	healthy := new(userV029)
	_, err = engine.Where("login = ?", "healthy").Get(healthy)
	require.NoError(t, err)
	assert.EqualValues(t, 1, healthy.ForgeID)

	otherOrg := new(orgV029)
	_, err = engine.Where("name = ?", "other-forge").Get(otherOrg)
	require.NoError(t, err)
	assert.EqualValues(t, 2, otherOrg.ForgeID)
}
