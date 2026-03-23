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
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
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

// buildUnsignedToken manually constructs a JWT with alg=none so we can verify
// that Verify() rejects it even though the signature section is empty.
// We do NOT use the golang-jwt library here because modern versions refuse to
// produce none-signed tokens — that is exactly the property we want to test.
func buildUnsignedToken(t *testing.T, agentID int64) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString(
		jwtMustMarshal(t, map[string]string{"alg": "none", "typ": "JWT"}),
	)
	payload := base64.RawURLEncoding.EncodeToString(
		jwtMustMarshal(t, map[string]any{
			"agent_id": agentID,
			"iss":      "woodpecker",
		}),
	)
	// A none-signed JWT carries an empty signature segment.
	return header + "." + payload + "."
}

// buildRS256FakeToken constructs a JWT header claiming RS256 to exercise the
// unexpected-signing-method guard inside JWTManager.Verify().
func buildRS256FakeToken(t *testing.T) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString(
		jwtMustMarshal(t, map[string]string{"alg": "RS256", "typ": "JWT"}),
	)
	payload := base64.RawURLEncoding.EncodeToString(
		jwtMustMarshal(t, map[string]any{"agent_id": 1, "iss": "woodpecker"}),
	)
	sig := base64.RawURLEncoding.EncodeToString([]byte("fake-rsa-sig"))
	return header + "." + payload + "." + sig
}

// buildFutureNbfToken constructs a JWT whose nbf claim is set far in the
// future. The token must be rejected regardless of which check fires first.
func buildFutureNbfToken(t *testing.T) string {
	t.Helper()
	const farFuture = int64(9_999_999_999) // year 2286
	header := base64.RawURLEncoding.EncodeToString(
		jwtMustMarshal(t, map[string]string{"alg": "HS256", "typ": "JWT"}),
	)
	payload := base64.RawURLEncoding.EncodeToString(
		jwtMustMarshal(t, map[string]any{
			"agent_id": 1,
			"iss":      "woodpecker",
			"nbf":      farFuture,
			"exp":      farFuture + 3600,
		}),
	)
	badSig := base64.RawURLEncoding.EncodeToString([]byte("bad"))
	return header + "." + payload + "." + badSig
}

func jwtMustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return b
}

func TestJWTManagerAdditional(t *testing.T) {
	t.Parallel()

	t.Run("none-algorithm token is rejected", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")
		noneToken := buildUnsignedToken(t, 42)

		// Sanity: token really does carry the none algorithm header.
		parts := strings.Split(noneToken, ".")
		require.Len(t, parts, 3)
		assert.Equal(t, "", parts[2], "signature part must be empty for none-alg tokens")

		_, err := manager.Verify(noneToken)
		assert.Error(t, err, "verifier must reject a none-algorithm token")
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("RS256 token (unexpected signing method) is rejected", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")
		rs256Token := buildRS256FakeToken(t)

		_, err := manager.Verify(rs256Token)
		assert.Error(t, err, "verifier must reject tokens with an unexpected signing method")
		assert.Contains(t, err.Error(), "invalid token")
	})

	t.Run("token with far-future NotBefore is rejected", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")
		futureToken := buildFutureNbfToken(t)

		_, err := manager.Verify(futureToken)
		assert.Error(t, err)
	})

	t.Run("two valid tokens for same agent are each independently verifiable", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")

		tok1, err := manager.Generate(5)
		require.NoError(t, err)
		tok2, err := manager.Generate(5)
		require.NoError(t, err)

		claims1, err := manager.Verify(tok1)
		require.NoError(t, err)
		assert.Equal(t, int64(5), claims1.AgentID)

		claims2, err := manager.Verify(tok2)
		require.NoError(t, err)
		assert.Equal(t, int64(5), claims2.AgentID)
	})

	t.Run("zero agent ID is preserved through generate/verify roundtrip", func(t *testing.T) {
		t.Parallel()

		manager := NewJWTManager("test-secret")
		token, err := manager.Generate(0)
		require.NoError(t, err)

		claims, err := manager.Verify(token)
		require.NoError(t, err)
		assert.Equal(t, int64(0), claims.AgentID)
	})
}
