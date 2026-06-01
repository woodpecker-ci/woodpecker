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
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetConnectionRetryTimeout(t *testing.T) {
	tc := []struct {
		name    string
		timeout time.Duration
	}{
		{"finite", 5 * time.Minute},
		{"zero means infinite", 0},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			cl := &client{}
			SetConnectionRetryTimeout(c.timeout)(cl)
			assert.Equal(t, c.timeout, cl.connectionRetryTimeout)
		})
	}
}

func TestIsConnected(t *testing.T) {
	cl := &client{conn: newTestConn(t)}
	defer cl.conn.Close()

	t.Run("idle connection reports connected", func(t *testing.T) {
		assert.True(t, cl.IsConnected())
	})

	t.Run("closed connection reports not connected", func(t *testing.T) {
		assert.NoError(t, cl.conn.Close())
		assert.False(t, cl.IsConnected())
	})
}
