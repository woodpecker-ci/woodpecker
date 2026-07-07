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
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-github/v88/github"
)

const (
	// Lifetime of the JWT used to authenticate as the GitHub App itself.
	// GitHub allows at most 10 minutes.
	appJWTLifetime = 5 * time.Minute
	// Subtracted from the JWT issued-at time to protect against clock drift
	// between Woodpecker and GitHub, as recommended by GitHub.
	appJWTClockDrift = 60 * time.Second
	// How long a resolved repository -> installation mapping is cached.
	installationTTL = time.Hour
	// How long an "app is not installed on this repository" result is
	// cached, so newly installed apps are picked up reasonably fast while not
	// looking up the installation again on every single API call for
	// repositories the app does not cover.
	installationNotInstalledTTL = 5 * time.Minute
	// Minimum remaining installation token validity required for server-side
	// API calls.
	apiTokenMinValidity = 5 * time.Minute
	// Minimum remaining installation token validity required for tokens
	// handed out as clone credentials. Installation tokens live for one hour
	// and the clone happens on an agent at some later point in time, so keep
	// the margin large.
	netrcTokenMinValidity = 45 * time.Minute
	// Bounds the token mint call done by Netrc, which does not get a context
	// from the forge interface.
	netrcTokenTimeout = 10 * time.Second
)

// errAppNotInstalled is returned when the GitHub App is not installed on the
// requested repository. Callers fall back to user OAuth tokens.
var errAppNotInstalled = errors.New("github app is not installed on this repository")

// appClient authenticates as a GitHub App and mints installation access
// tokens which are used for server-side API calls and as clone credentials.
// Installation lookups and tokens are cached; a cached token is reused as
// long as it satisfies the minimum remaining validity requested by the
// caller. Concurrent cache misses may mint duplicate tokens, which is
// harmless as GitHub allows multiple active installation tokens.
type appClient struct {
	appID      string // app id or client id of the GitHub App
	privateKey *rsa.PrivateKey
	api        string
	skipVerify bool

	mu            sync.Mutex
	installations map[string]installationEntry // repo full name -> installation
	tokens        map[int64]tokenEntry         // installation id -> access token
	cloneTokens   map[string]tokenEntry        // repo full name -> repo-scoped clone token
	now           func() time.Time             // overridable for tests
}

type installationEntry struct {
	id      int64 // zero when the app is not installed on the repository
	validTo time.Time
}

type tokenEntry struct {
	token     string
	expiresAt time.Time
}

func newAppClient(appID, privateKey, api string, skipVerify bool) (*appClient, error) {
	key, err := parseAppPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse github app private key: %w", err)
	}
	return &appClient{
		appID:         appID,
		privateKey:    key,
		api:           api,
		skipVerify:    skipVerify,
		installations: make(map[string]installationEntry),
		tokens:        make(map[int64]tokenEntry),
		cloneTokens:   make(map[string]tokenEntry),
		now:           time.Now,
	}, nil
}

// parseAppPrivateKey parses the RSA private key of the GitHub App, accepting
// both plain PEM and base64-encoded PEM.
func parseAppPrivateKey(key string) (*rsa.PrivateKey, error) {
	raw := strings.TrimSpace(key)
	if !strings.Contains(raw, "-----BEGIN") {
		decoded, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(raw), ""))
		if err != nil {
			return nil, fmt.Errorf("private key is neither PEM nor base64-encoded PEM: %w", err)
		}
		raw = string(decoded)
	}

	parsed, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(raw))
	if err != nil {
		// keys pasted into single-line inputs or env vars often lose their
		// newlines, which makes the PEM unparsable - restore them
		if repaired, ok := repairSquashedPEM(raw); ok {
			if parsed, repairErr := jwt.ParseRSAPrivateKeyFromPEM([]byte(repaired)); repairErr == nil {
				return parsed, nil
			}
		}
		return nil, err
	}
	return parsed, nil
}

var squashedPEMRegex = regexp.MustCompile(`(?s)^-----BEGIN ([A-Z0-9 ]+)-----\s*(.+?)\s*-----END ([A-Z0-9 ]+)-----$`)

// repairSquashedPEM restores the line structure of a PEM block whose
// newlines were stripped, e.g. by pasting it into a single-line input field.
func repairSquashedPEM(raw string) (string, bool) {
	match := squashedPEMRegex.FindStringSubmatch(strings.TrimSpace(raw))
	if match == nil || match[1] != match[3] {
		return "", false
	}
	body := strings.Join(strings.Fields(match[2]), "")
	var pem strings.Builder
	pem.WriteString("-----BEGIN " + match[1] + "-----\n")
	const lineLength = 64
	for i := 0; i < len(body); i += lineLength {
		pem.WriteString(body[i:min(i+lineLength, len(body))] + "\n")
	}
	pem.WriteString("-----END " + match[1] + "-----\n")
	return pem.String(), true
}

// appJWT returns a short-lived JWT that authenticates requests as the GitHub
// App itself, as required by the installation endpoints.
func (a *appClient) appJWT() (string, error) {
	now := a.now()
	claims := jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(now.Add(-appJWTClockDrift)),
		ExpiresAt: jwt.NewNumericDate(now.Add(appJWTLifetime)),
		Issuer:    a.appID,
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(a.privateKey)
}

// newAppJWTClient returns a client authenticating as the GitHub App itself.
func (a *appClient) newAppJWTClient(ctx context.Context) (*github.Client, error) {
	appJWT, err := a.appJWT()
	if err != nil {
		return nil, fmt.Errorf("failed to sign github app JWT: %w", err)
	}
	return newGithubClient(ctx, a.api, appJWT, a.skipVerify)
}

func installationKey(owner, name string) string {
	return strings.ToLower(owner + "/" + name)
}

// cacheNotInstalled remembers for a limited time that the GitHub App cannot
// serve the given repository, so callers fall back to user tokens without
// looking up the installation again on every call.
func (a *appClient) cacheNotInstalled(owner, name string) {
	a.mu.Lock()
	a.installations[installationKey(owner, name)] = installationEntry{id: 0, validTo: a.now().Add(installationNotInstalledTTL)}
	a.mu.Unlock()
}

// installationID resolves the installation of the GitHub App covering the
// given repository, returning errAppNotInstalled when there is none.
func (a *appClient) installationID(ctx context.Context, owner, name string) (int64, error) {
	key := installationKey(owner, name)

	a.mu.Lock()
	entry, ok := a.installations[key]
	a.mu.Unlock()
	if ok && a.now().Before(entry.validTo) {
		if entry.id == 0 {
			return 0, errAppNotInstalled
		}
		return entry.id, nil
	}

	client, err := a.newAppJWTClient(ctx)
	if err != nil {
		return 0, err
	}

	installation, resp, err := client.Apps.GetRepositoryInstallation(ctx, owner, name)
	if resp != nil && resp.StatusCode == http.StatusNotFound {
		a.cacheNotInstalled(owner, name)
		return 0, errAppNotInstalled
	}
	if err != nil {
		return 0, fmt.Errorf("failed to look up github app installation for %s/%s: %w", owner, name, err)
	}

	id := installation.GetID()
	a.mu.Lock()
	a.installations[key] = installationEntry{id: id, validTo: a.now().Add(installationTTL)}
	a.mu.Unlock()
	return id, nil
}

// token returns an installation access token for the repository that is
// valid for at least minValidity. The token covers the whole installation
// with the permissions granted to the app, and is used for server-side API
// calls only.
func (a *appClient) token(ctx context.Context, owner, name string, minValidity time.Duration) (string, error) {
	installationID, err := a.installationID(ctx, owner, name)
	if err != nil {
		return "", err
	}

	a.mu.Lock()
	entry, ok := a.tokens[installationID]
	a.mu.Unlock()
	if ok && a.now().Add(minValidity).Before(entry.expiresAt) {
		return entry.token, nil
	}

	token, err := a.mintToken(ctx, installationID, owner, name, nil)
	if err != nil {
		return "", err
	}

	a.mu.Lock()
	a.tokens[installationID] = *token
	a.mu.Unlock()
	return token.token, nil
}

// cloneToken returns an installation access token for the repository that is
// restricted to the repository itself and read-only contents access. It is
// handed out as clone credential, so it carries as little access as
// possible. When wideScope is true the token covers the whole installation
// with the app's permissions instead, so that clones of other repositories
// of the same installation (e.g. git submodules or private go modules) keep
// working.
func (a *appClient) cloneToken(ctx context.Context, owner, name string, minValidity time.Duration, wideScope bool) (string, error) {
	if wideScope {
		return a.token(ctx, owner, name, minValidity)
	}

	installationID, err := a.installationID(ctx, owner, name)
	if err != nil {
		return "", err
	}

	key := installationKey(owner, name)
	a.mu.Lock()
	entry, ok := a.cloneTokens[key]
	a.mu.Unlock()
	if ok && a.now().Add(minValidity).Before(entry.expiresAt) {
		return entry.token, nil
	}

	token, err := a.mintToken(ctx, installationID, owner, name, &github.InstallationTokenOptions{
		Repositories: []string{name},
		Permissions: &github.InstallationPermissions{
			Contents: github.Ptr("read"),
		},
	})
	if err != nil {
		return "", err
	}

	a.mu.Lock()
	a.cloneTokens[key] = *token
	a.mu.Unlock()
	return token.token, nil
}

// mintToken creates a new installation access token, optionally restricted
// to specific repositories and permissions.
func (a *appClient) mintToken(ctx context.Context, installationID int64, owner, name string, opts *github.InstallationTokenOptions) (*tokenEntry, error) {
	client, err := a.newAppJWTClient(ctx)
	if err != nil {
		return nil, err
	}

	token, resp, err := client.Apps.CreateInstallationToken(ctx, installationID, opts)
	if resp != nil && (resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden) {
		// the installation was removed or suspended after it was cached,
		// fall back to user tokens until the app becomes available again
		a.cacheNotInstalled(owner, name)
		return nil, errAppNotInstalled
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create github app installation token for %s/%s: %w", owner, name, err)
	}

	return &tokenEntry{token: token.GetToken(), expiresAt: token.GetExpiresAt().Time}, nil
}
