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

package github

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateAppKey(t *testing.T) (*rsa.PrivateKey, string) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pemKey := string(pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}))
	return key, pemKey
}

func TestParseAppPrivateKey(t *testing.T) {
	key, pemKey := generateAppKey(t)

	t.Run("PKCS1 PEM", func(t *testing.T) {
		parsed, err := parseAppPrivateKey(pemKey)
		require.NoError(t, err)
		assert.True(t, key.Equal(parsed))
	})

	t.Run("PKCS8 PEM", func(t *testing.T) {
		pkcs8, err := x509.MarshalPKCS8PrivateKey(key)
		require.NoError(t, err)
		parsed, err := parseAppPrivateKey(string(pem.EncodeToMemory(&pem.Block{
			Type:  "PRIVATE KEY",
			Bytes: pkcs8,
		})))
		require.NoError(t, err)
		assert.True(t, key.Equal(parsed))
	})

	t.Run("base64-encoded PEM", func(t *testing.T) {
		parsed, err := parseAppPrivateKey(base64.StdEncoding.EncodeToString([]byte(pemKey)))
		require.NoError(t, err)
		assert.True(t, key.Equal(parsed))
	})

	t.Run("base64-encoded PEM with line breaks", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte(pemKey))
		var wrapped strings.Builder
		for i := 0; i < len(encoded); i += 64 {
			wrapped.WriteString(encoded[i:min(i+64, len(encoded))] + "\n")
		}
		parsed, err := parseAppPrivateKey(wrapped.String())
		require.NoError(t, err)
		assert.True(t, key.Equal(parsed))
	})

	t.Run("PEM with stripped newlines", func(t *testing.T) {
		// pasting a PEM file into a single-line input strips its newlines
		parsed, err := parseAppPrivateKey(strings.ReplaceAll(pemKey, "\n", ""))
		require.NoError(t, err)
		assert.True(t, key.Equal(parsed))
	})

	t.Run("base64-encoded PEM with stripped newlines", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte(strings.ReplaceAll(pemKey, "\n", "")))
		parsed, err := parseAppPrivateKey(encoded)
		require.NoError(t, err)
		assert.True(t, key.Equal(parsed))
	})

	t.Run("garbage", func(t *testing.T) {
		_, err := parseAppPrivateKey("not-a-key")
		assert.Error(t, err)
	})

	t.Run("squashed PEM with mismatched block types", func(t *testing.T) {
		_, ok := repairSquashedPEM("-----BEGIN RSA PRIVATE KEY-----abc-----END EC PRIVATE KEY-----")
		assert.False(t, ok)
	})

	t.Run("squashed PEM with invalid body", func(t *testing.T) {
		_, err := parseAppPrivateKey("-----BEGIN RSA PRIVATE KEY-----bm90LWEta2V5-----END RSA PRIVATE KEY-----")
		assert.Error(t, err)
	})
}

func TestAppJWT(t *testing.T) {
	key, pemKey := generateAppKey(t)

	app, err := newAppClient("12345", pemKey, defaultAPI, false)
	require.NoError(t, err)

	signed, err := app.appJWT()
	require.NoError(t, err)

	claims := &jwt.RegisteredClaims{}
	parsed, err := jwt.ParseWithClaims(signed, claims, func(_ *jwt.Token) (any, error) {
		return &key.PublicKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}))
	require.NoError(t, err)
	assert.True(t, parsed.Valid)
	assert.Equal(t, "12345", claims.Issuer)
	// issued-at must be backdated to protect against clock drift
	assert.WithinDuration(t, time.Now().Add(-appJWTClockDrift), claims.IssuedAt.Time, 5*time.Second)
	// GitHub rejects JWTs with a lifetime of more than 10 minutes
	assert.LessOrEqual(t, claims.ExpiresAt.Sub(claims.IssuedAt.Time), 10*time.Minute)
}

func TestAppInstallationToken(t *testing.T) {
	key, pemKey := generateAppKey(t)

	var installationLookups, tokenMints int
	verifyAppJWT := func(r *http.Request) error {
		appJWT := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		_, err := jwt.Parse(appJWT, func(_ *jwt.Token) (any, error) {
			return &key.PublicKey, nil
		}, jwt.WithValidMethods([]string{"RS256"}), jwt.WithIssuer("12345"))
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /repos/{owner}/{name}/installation", func(w http.ResponseWriter, r *http.Request) {
		installationLookups++
		if err := verifyAppJWT(r); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		switch r.PathValue("owner") {
		case "not-installed":
			http.Error(w, "", http.StatusNotFound)
		case "lookup-error":
			http.Error(w, "", http.StatusInternalServerError)
		case "uninstalled":
			// installation still resolves, but minting tokens for it fails
			fmt.Fprint(w, `{"id": 8}`)
		case "mint-error":
			fmt.Fprint(w, `{"id": 9}`)
		default:
			fmt.Fprint(w, `{"id": 7}`)
		}
	})
	mux.HandleFunc("POST /app/installations/{id}/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		tokenMints++
		if err := verifyAppJWT(r); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		switch r.PathValue("id") {
		case "8":
			http.Error(w, "", http.StatusNotFound)
			return
		case "9":
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, `{"token": "ghs_token_%d", "expires_at": %q}`,
			tokenMints, time.Now().Add(time.Hour).UTC().Format(time.RFC3339))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	app, err := newAppClient("12345", pemKey, server.URL+"/", false)
	require.NoError(t, err)
	now := time.Now()
	app.now = func() time.Time { return now }

	ctx := t.Context()

	token, err := app.token(ctx, "octocat", "hello-world", apiTokenMinValidity)
	require.NoError(t, err)
	assert.Equal(t, "ghs_token_1", token)
	assert.Equal(t, 1, installationLookups)
	assert.Equal(t, 1, tokenMints)

	// second call is served from the cache
	token, err = app.token(ctx, "octocat", "hello-world", apiTokenMinValidity)
	require.NoError(t, err)
	assert.Equal(t, "ghs_token_1", token)
	assert.Equal(t, 1, installationLookups)
	assert.Equal(t, 1, tokenMints)

	// another repository of the same installation shares the token
	token, err = app.token(ctx, "octocat", "other-repo", apiTokenMinValidity)
	require.NoError(t, err)
	assert.Equal(t, "ghs_token_1", token)
	assert.Equal(t, 2, installationLookups)
	assert.Equal(t, 1, tokenMints)

	// 30 minutes later the token is still fine for API calls but no longer
	// fresh enough to be handed out as clone credentials
	now = now.Add(30 * time.Minute)
	token, err = app.token(ctx, "octocat", "hello-world", apiTokenMinValidity)
	require.NoError(t, err)
	assert.Equal(t, "ghs_token_1", token)
	assert.Equal(t, 1, tokenMints)

	token, err = app.token(ctx, "octocat", "hello-world", netrcTokenMinValidity)
	require.NoError(t, err)
	assert.Equal(t, "ghs_token_2", token)
	assert.Equal(t, 2, tokenMints)

	// repositories the app is not installed on return errAppNotInstalled and
	// the negative result is cached
	lookups := installationLookups
	_, err = app.token(ctx, "not-installed", "some-repo", apiTokenMinValidity)
	assert.ErrorIs(t, err, errAppNotInstalled)
	assert.Equal(t, lookups+1, installationLookups)

	_, err = app.token(ctx, "not-installed", "some-repo", apiTokenMinValidity)
	assert.ErrorIs(t, err, errAppNotInstalled)
	assert.Equal(t, lookups+1, installationLookups)

	// ... but only for a limited time, so app installations are picked up
	now = now.Add(installationNotInstalledTTL + time.Second)
	_, err = app.token(ctx, "not-installed", "some-repo", apiTokenMinValidity)
	assert.ErrorIs(t, err, errAppNotInstalled)
	assert.Equal(t, lookups+2, installationLookups)

	// an installation that resolves but can no longer mint tokens (app was
	// uninstalled or suspended) falls back like a missing installation and
	// does not retry the doomed mint on the next call
	mints := tokenMints
	_, err = app.token(ctx, "uninstalled", "some-repo", apiTokenMinValidity)
	assert.ErrorIs(t, err, errAppNotInstalled)
	assert.Equal(t, mints+1, tokenMints)

	_, err = app.token(ctx, "uninstalled", "some-repo", apiTokenMinValidity)
	assert.ErrorIs(t, err, errAppNotInstalled)
	assert.Equal(t, mints+1, tokenMints)

	// transient installation lookup errors are surfaced, not cached
	lookups = installationLookups
	_, err = app.token(ctx, "lookup-error", "some-repo", apiTokenMinValidity)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, errAppNotInstalled)

	_, err = app.token(ctx, "lookup-error", "some-repo", apiTokenMinValidity)
	assert.Error(t, err)
	assert.Equal(t, lookups+2, installationLookups)

	// transient mint errors are surfaced and retried on the next call
	mints = tokenMints
	_, err = app.token(ctx, "mint-error", "some-repo", apiTokenMinValidity)
	assert.Error(t, err)
	assert.NotErrorIs(t, err, errAppNotInstalled)

	_, err = app.token(ctx, "mint-error", "some-repo", apiTokenMinValidity)
	assert.Error(t, err)
	assert.Equal(t, mints+2, tokenMints)
}
