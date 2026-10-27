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

type userV031 struct {
	ID            int64  `xorm:"pk autoincr 'id'"`
	ForgeID       int64  `xorm:"forge_id"`
	ForgeRemoteID string `xorm:"forge_remote_id"`
	Login         string `xorm:"'login'"`
	OrgID         int64  `xorm:"org_id"`
}

func (userV031) TableName() string { return "users" }

func TestSetOrgForgeRemoteIDFromUsers(t *testing.T) {
	engine, closeDB := testDB(t, true)
	defer closeDB()

	require.NoError(t, engine.Sync(new(userV031), new(orgV029)))

	personal := &orgV029{ForgeID: 1, Name: "alice", IsUser: true}
	noRemoteID := &orgV029{ForgeID: 1, Name: "bob", IsUser: true}
	team := &orgV029{ForgeID: 1, Name: "team"}
	for _, o := range []*orgV029{personal, noRemoteID, team} {
		_, err := engine.Insert(o)
		require.NoError(t, err)
	}
	_, err := engine.Insert(&userV031{ForgeID: 1, ForgeRemoteID: "7", Login: "alice", OrgID: personal.ID})
	require.NoError(t, err)
	// users created by an admin have no forge remote id yet
	_, err = engine.Insert(&userV031{ForgeID: 1, ForgeRemoteID: "0", Login: "bob", OrgID: noRemoteID.ID})
	require.NoError(t, err)

	sess := engine.NewSession()
	defer sess.Close()
	require.NoError(t, setOrgForgeRemoteIDFromUsers.MigrateSession(sess))
	require.NoError(t, sess.Commit())

	remoteIDs := map[int64]string{}
	var orgs []*orgV031
	require.NoError(t, engine.Find(&orgs))
	for _, o := range orgs {
		remoteIDs[o.ID] = o.ForgeRemoteID
	}
	assert.Equal(t, map[int64]string{personal.ID: "7", noRemoteID.ID: "", team.ID: ""}, remoteIDs)
}
