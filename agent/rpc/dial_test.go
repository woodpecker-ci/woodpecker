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

package rpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func newTestConn(t *testing.T) *grpc.ClientConn {
	t.Helper()
	conn, err := grpc.NewClient("localhost:0", grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	return conn
}

func TestAgentConnClose(t *testing.T) {
	t.Run("both nil is safe", func(t *testing.T) {
		c := &AgentConn{}
		assert.NotPanics(t, c.Close)
	})

	t.Run("only auth set", func(t *testing.T) {
		c := &AgentConn{AuthConn: newTestConn(t)}
		assert.NotPanics(t, c.Close)
	})

	t.Run("only main set", func(t *testing.T) {
		c := &AgentConn{MainConn: newTestConn(t)}
		assert.NotPanics(t, c.Close)
	})

	t.Run("both set closes both", func(t *testing.T) {
		c := &AgentConn{
			AuthConn: newTestConn(t),
			MainConn: newTestConn(t),
		}
		assert.NotPanics(t, c.Close)
	})
}
