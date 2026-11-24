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

//go:build test

package exec

import (
	"bytes"
	"crypto/ed25519"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaronf/httpsign"
	"github.com/zalando/go-keyring"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/builder"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
)

func TestMain(m *testing.M) {
	// Keep the signing key in memory. The real keyring would prompt or persist
	// a key on the machine running the tests.
	keyring.MockInit()
	os.Exit(m.Run())
}

const extensionWorkflow = `when:
  - event: manual

steps:
  - name: build
    image: alpine
    commands:
      - echo from-extension
`

func TestFetchConfigExtension(t *testing.T) {
	m := &metadata.Metadata{
		Repo: metadata.Repo{Owner: "octocat", Name: "hello-world", Branch: "main"},
		Curr: metadata.Pipeline{
			Event:  metadata.EventManual,
			Commit: metadata.Commit{Sha: "abc123", Branch: "feature", Ref: "refs/heads/feature"},
		},
	}
	local := []*builder.YamlFile{{Name: ".woodpecker.yaml", Data: []byte("steps: {}\n")}}

	t.Run("ReturnsConfigs", func(t *testing.T) {
		privateKey, err := configExtensionPrivateKey()
		require.NoError(t, err)
		publicKey, ok := privateKey.Public().(ed25519.PublicKey)
		require.True(t, ok)

		var got map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Contains(t, r.Header.Get("User-Agent"), "Woodpecker/")
			verifier, verifyErr := httpsign.NewEd25519Verifier(publicKey,
				httpsign.NewVerifyConfig(),
				httpsign.Headers("@request-target", "content-digest"))
			assert.NoError(t, verifyErr)
			assert.NoError(t, httpsign.VerifyRequest(configExtensionSignatureName, *verifier, r))
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
			assert.NoError(t, json.NewEncoder(w).Encode(configExtensionResponse{
				Configs: []*configExtensionFile{{Name: "build", Data: extensionWorkflow}},
			}))
		}))
		t.Cleanup(srv.Close)

		yamls, err := fetchConfigExtension(t.Context(), srv.URL, m, local)
		require.NoError(t, err)
		require.Len(t, yamls, 1)
		assert.Equal(t, "build", yamls[0].Name)
		assert.Equal(t, extensionWorkflow, string(yamls[0].Data))

		// The field names must match what the server sends.
		repo, ok := got["repo"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "octocat", repo["owner"])
		assert.Equal(t, "hello-world", repo["name"])
		assert.Equal(t, "octocat/hello-world", repo["full_name"])
		assert.Equal(t, "main", repo["default_branch"])
		pipeline, ok := got["pipeline"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "manual", pipeline["event"])
		assert.Equal(t, "abc123", pipeline["commit"])
		assert.Equal(t, "feature", pipeline["branch"])
		assert.Equal(t, "refs/heads/feature", pipeline["ref"])
		assert.Equal(t, false, got["restart"])
		assert.Equal(t, []any{map[string]any{"name": ".woodpecker.yaml", "data": "steps: {}\n"}}, got["configuration"])
	})

	t.Run("NoContentUsesLocal", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		t.Cleanup(srv.Close)

		yamls, err := fetchConfigExtension(t.Context(), srv.URL, m, local)
		require.NoError(t, err)
		assert.Equal(t, local, yamls)
	})

	t.Run("NoContentWithoutLocal", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		t.Cleanup(srv.Close)

		_, err := fetchConfigExtension(t.Context(), srv.URL, m, nil)
		assert.ErrorContains(t, err, "no local workflow files")
	})

	t.Run("ErrorStatus", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "language not supported", http.StatusNotImplemented)
		}))
		t.Cleanup(srv.Close)

		_, err := fetchConfigExtension(t.Context(), srv.URL, m, local)
		assert.ErrorContains(t, err, "status 501: language not supported")
	})

	t.Run("NoConfigs", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, err := w.Write([]byte(`{"configs":[]}`))
			assert.NoError(t, err)
		}))
		t.Cleanup(srv.Close)

		_, err := fetchConfigExtension(t.Context(), srv.URL, m, local)
		assert.ErrorContains(t, err, "returned no configs")
	})

	t.Run("NotJSON", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, err := w.Write([]byte("<!doctype html>"))
			assert.NoError(t, err)
		}))
		t.Cleanup(srv.Close)

		_, err := fetchConfigExtension(t.Context(), srv.URL, m, local)
		assert.ErrorContains(t, err, "could not decode config extension response")
	})
}

func TestExecConfigExtension(t *testing.T) {
	t.Run("RunsExtensionWorkflow", func(t *testing.T) {
		repoDir := t.TempDir()

		var got map[string]any
		var signatureInput string
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			signatureInput = r.Header.Get("Signature-Input")
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
			assert.NoError(t, json.NewEncoder(w).Encode(configExtensionResponse{
				Configs: []*configExtensionFile{{Name: "build", Data: extensionWorkflow}},
			}))
		}))
		t.Cleanup(srv.Close)

		clearEnv(t)
		output := captureStderr(t, func() {
			require.NoError(t, Command.Run(t.Context(), []string{
				"woodpecker-cli",
				"--backend-engine", "dummy",
				"--repo-path", repoDir,
				"--repo", "octocat/hello-world",
				"--config-extension-endpoint", srv.URL,
			}))
		})

		assert.Contains(t, output, "echo from-extension")
		assert.Contains(t, output, "BEGIN PUBLIC KEY")
		assert.Contains(t, signatureInput, configExtensionSignatureName)
		repo, ok := got["repo"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "octocat/hello-world", repo["full_name"])
		assert.NotContains(t, got, "configuration", "no local workflow files, so none are sent")
	})

	t.Run("NoContentRunsLocalWorkflow", func(t *testing.T) {
		repoDir := t.TempDir()
		localWorkflow := `when:
  - event: manual

steps:
  - name: build
    image: alpine
    commands:
      - echo from-local
`
		require.NoError(t, os.WriteFile(filepath.Join(repoDir, ".woodpecker.yaml"), []byte(localWorkflow), 0o600))

		var got map[string]any
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.NoError(t, json.NewDecoder(r.Body).Decode(&got))
			w.WriteHeader(http.StatusNoContent)
		}))
		t.Cleanup(srv.Close)

		clearEnv(t)
		output := captureStderr(t, func() {
			require.NoError(t, Command.Run(t.Context(), []string{
				"woodpecker-cli",
				"--backend-engine", "dummy",
				"--repo-path", repoDir,
				"--config-extension-endpoint", srv.URL,
			}))
		})

		assert.Contains(t, output, "echo from-local")
		assert.Equal(t, []any{map[string]any{"name": ".woodpecker.yaml", "data": localWorkflow}}, got["configuration"])
	})
}

func TestConfigExtensionPrivateKey(t *testing.T) {
	t.Run("ReusesStoredKey", func(t *testing.T) {
		first, err := configExtensionPrivateKey()
		require.NoError(t, err)
		second, err := configExtensionPrivateKey()
		require.NoError(t, err)
		assert.Equal(t, []byte(first), []byte(second))
	})

	t.Run("RejectsCorruptStoredKey", func(t *testing.T) {
		require.NoError(t, keyring.Set(configExtensionKeyringService, configExtensionKeyringUser, "not-hex"))
		t.Cleanup(func() {
			assert.NoError(t, keyring.Delete(configExtensionKeyringService, configExtensionKeyringUser))
		})

		_, err := configExtensionPrivateKey()
		assert.ErrorContains(t, err, "could not decode the stored config extension signing key")
	})

	t.Run("PublicKeyMatchesServerFormat", func(t *testing.T) {
		privateKey, err := configExtensionPrivateKey()
		require.NoError(t, err)
		encoded, err := configExtensionPublicKeyPEM(privateKey)
		require.NoError(t, err)

		block, rest := pem.Decode([]byte(encoded))
		require.NotNil(t, block)
		assert.Empty(t, rest)
		assert.Equal(t, "PUBLIC KEY", block.Type)
		parsed, err := x509.ParsePKIXPublicKey(block.Bytes)
		require.NoError(t, err)
		assert.Equal(t, privateKey.Public(), parsed)
	})
}

func TestRelativeConfigName(t *testing.T) {
	repo := t.TempDir()
	assert.Equal(t, ".woodpecker/build.yaml", relativeConfigName(repo, filepath.Join(repo, ".woodpecker", "build.yaml")))
	outside := filepath.Join(filepath.Dir(repo), "other.yaml")
	assert.Equal(t, outside, relativeConfigName(repo, outside))
}

// captureStderr returns what fn writes to os.Stderr, where the step output goes.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	defer func() { os.Stderr = oldStderr }()

	done := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, copyErr := io.Copy(&buf, r)
		assert.NoError(t, copyErr)
		done <- buf.String()
	}()

	fn()
	require.NoError(t, w.Close())
	output := <-done
	require.NoError(t, r.Close())
	return output
}
