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

	"go.woodpecker-ci.org/woodpecker/v3/server/model"
)

type agentV031 struct {
	ID    int64  `xorm:"pk autoincr 'id'"`
	Token string `xorm:"token"`
}

func (agentV031) TableName() string { return "agents" }

type agentV032 struct {
	ID        int64  `xorm:"pk autoincr 'id'"`
	Token     string `xorm:"token"`
	TokenHash string `xorm:"token_hash"`
}

func (agentV032) TableName() string { return "agents" }

func TestAddAgentTokenHash(t *testing.T) {
	engine, closeDB := testDB(t, true)
	defer closeDB()

	require.NoError(t, engine.Sync(new(agentV031)))
	_, err := engine.Insert(&agentV031{Token: "token-a"}, &agentV031{Token: "token-b"})
	require.NoError(t, err)

	sess := engine.NewSession()
	defer sess.Close()
	require.NoError(t, addAgentTokenHash.MigrateSession(sess))
	require.NoError(t, sess.Commit())

	var agents []*agentV032
	require.NoError(t, engine.OrderBy("id").Find(&agents))
	require.Len(t, agents, 2)
	for _, agent := range agents {
		assert.Equal(t, model.HashAgentToken(agent.Token), agent.TokenHash)
	}
}
