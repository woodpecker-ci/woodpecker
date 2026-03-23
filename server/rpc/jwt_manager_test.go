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

package rpc

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTManager(t *testing.T) {
	t.Parallel()

	t.Run("generate and verify roundtrip", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")
		token, err := manager.Generate(42)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := manager.Verify(token)
		require.NoError(t, err)
		assert.Equal(t, int64(42), claims.AgentID)
	})

	t.Run("claims contain correct fields", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")
		token, err := manager.Generate(99)
		require.NoError(t, err)

		claims, err := manager.Verify(token)
		require.NoError(t, err)

		assert.Equal(t, int64(99), claims.AgentID)
		assert.Equal(t, "woodpecker", claims.Issuer)
		assert.Equal(t, fmt.Sprintf("%d", 99), claims.Subject)
		assert.Equal(t, fmt.Sprintf("%d", 99), claims.ID)
	})

	t.Run("different agent IDs produce different tokens", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")
		token1, err := manager.Generate(1)
		require.NoError(t, err)

		token2, err := manager.Generate(2)
		require.NoError(t, err)

		assert.NotEqual(t, token1, token2)
	})

	t.Run("expired token is rejected", func(t *testing.T) {
		t.Parallel()

		manager := &JWTManager{
			secretKey:     "test-secret",
			tokenDuration: 1 * time.Millisecond,
		}

		token, err := manager.Generate(42)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		_, err = manager.Verify(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("wrong signing secret rejects token", func(t *testing.T) {
		t.Parallel()

		managerA := NewJWTManager("secret-A")
		managerB := NewJWTManager("secret-B")

		token, err := managerA.Generate(42)
		require.NoError(t, err)

		_, err = managerB.Verify(token)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("tampered token is rejected", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")
		token, err := manager.Generate(42)
		require.NoError(t, err)

		// flip a character in the signature portion
		tampered := token[:len(token)-1] + "X"

		_, err = manager.Verify(tampered)
		assert.Error(t, err)
	})

	t.Run("empty token is rejected", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")
		_, err := manager.Verify("")
		assert.Error(t, err)
	})

	t.Run("garbage token is rejected", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")
		_, err := manager.Verify("this-is-not-a-jwt")
		assert.Error(t, err)
	})

	t.Run("token generated with negative agent ID", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")
		token, err := manager.Generate(-1)
		require.NoError(t, err)

		claims, err := manager.Verify(token)
		require.NoError(t, err)
		assert.Equal(t, int64(-1), claims.AgentID)
	})
}
