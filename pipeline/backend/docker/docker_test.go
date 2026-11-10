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

package docker

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
)

type testCert struct {
	cert *x509.Certificate
	key  *ecdsa.PrivateKey
}

// newTestCert creates a certificate signed by parent, or a self-signed CA if parent is nil.
func newTestCert(t *testing.T, tmpl *x509.Certificate, parent *testCert) *testCert {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	tmpl.SerialNumber = big.NewInt(time.Now().UnixNano())
	tmpl.NotBefore = time.Now().Add(-time.Minute)
	tmpl.NotAfter = time.Now().Add(time.Hour)
	signer, signerKey := tmpl, key
	if parent == nil {
		tmpl.IsCA = true
		tmpl.BasicConstraintsValid = true
		tmpl.KeyUsage = x509.KeyUsageCertSign
	} else {
		signer, signerKey = parent.cert, parent.key
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, signer, &key.PublicKey, signerKey)
	require.NoError(t, err)
	cert, err := x509.ParseCertificate(der)
	require.NoError(t, err)
	return &testCert{cert: cert, key: key}
}

func (c *testCert) tlsCertificate() tls.Certificate {
	return tls.Certificate{Certificate: [][]byte{c.cert.Raw}, PrivateKey: c.key}
}

// writeTestCerts writes ca.pem, cert.pem and key.pem for a client certificate signed by ca into a new dir.
func writeTestCerts(t *testing.T, ca *testCert) string {
	t.Helper()

	clientCert := newTestCert(t, &x509.Certificate{
		Subject:     pkix.Name{CommonName: "woodpecker-agent"},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}, ca)
	keyDER, err := x509.MarshalECPrivateKey(clientCert.key)
	require.NoError(t, err)

	dir := t.TempDir()
	write := func(name, typ string, der []byte) {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}), 0o600))
	}
	write("ca.pem", "CERTIFICATE", ca.cert.Raw)
	write("cert.pem", "CERTIFICATE", clientCert.cert.Raw)
	write("key.pem", "EC PRIVATE KEY", keyDER)
	return dir
}

// fakeDaemon answers the requests docker.Load sends and records what it received.
type fakeDaemon struct {
	mu          sync.Mutex
	userAgents  []string
	clientCerts []string
}

func (d *fakeDaemon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	d.mu.Lock()
	d.userAgents = append(d.userAgents, r.UserAgent())
	if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
		d.clientCerts = append(d.clientCerts, r.TLS.PeerCertificates[0].Subject.CommonName)
	}
	d.mu.Unlock()

	w.Header().Set("Api-Version", "1.51")
	if strings.HasSuffix(r.URL.Path, "/info") {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"OSType":"linux","Architecture":"x86_64"}`))
		return
	}
	_, _ = w.Write([]byte("OK"))
}

// newTLSDaemon starts a fake daemon that requires a client certificate signed by clientCA
// and presents a server certificate signed by serverCA.
func newTLSDaemon(t *testing.T, serverCA, clientCA *testCert) (*fakeDaemon, string) {
	t.Helper()

	serverCert := newTestCert(t, &x509.Certificate{
		Subject:     pkix.Name{CommonName: "docker"},
		IPAddresses: []net.IP{net.ParseIP("127.0.0.1")},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}, serverCA)
	clientCAs := x509.NewCertPool()
	clientCAs.AddCert(clientCA.cert)

	daemon := &fakeDaemon{}
	srv := httptest.NewUnstartedServer(daemon)
	srv.TLS = &tls.Config{
		Certificates: []tls.Certificate{serverCert.tlsCertificate()},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCAs,
	}
	srv.StartTLS()
	t.Cleanup(srv.Close)
	return daemon, "tcp://" + srv.Listener.Addr().String()
}

// load runs docker.Load with the given agent arguments.
func load(t *testing.T, args ...string) (*backend_types.BackendInfo, error) {
	t.Helper()

	// the flags also read the docker environment variables, which must not leak into the test
	for _, env := range []string{"DOCKER_HOST", "DOCKER_CERT_PATH", "DOCKER_TLS_VERIFY", "DOCKER_API_VERSION"} {
		t.Setenv(env, "")
		require.NoError(t, os.Unsetenv(env))
	}

	var info *backend_types.BackendInfo
	cmd := &cli.Command{
		Flags: Flags,
		Action: func(ctx context.Context, c *cli.Command) error {
			var err error
			info, err = New().Load(context.WithValue(ctx, backend_types.CliCommand, c))
			return err
		},
	}
	err := cmd.Run(t.Context(), append([]string{"agent"}, args...))
	return info, err
}

func TestHTTPClientOfOptsWithHost(t *testing.T) {
	ca := newTestCert(t, &x509.Certificate{Subject: pkix.Name{CommonName: "ca"}}, nil)

	httpClient, err := httpClientOfOpts(writeTestCerts(t, ca), true)
	require.NoError(t, err)
	require.NotNil(t, httpClient)

	// client.WithHost has to configure the transport of the TLS client for the given host.
	_, err = client.New(client.WithHTTPClient(httpClient), client.WithHost("tcp://127.0.0.1:2376"))
	assert.NoError(t, err)
}

func TestLoad(t *testing.T) {
	userAgent := httputil.UserAgent("backend-docker")

	t.Run("tls with custom host", func(t *testing.T) {
		ca := newTestCert(t, &x509.Certificate{Subject: pkix.Name{CommonName: "ca"}}, nil)
		daemon, host := newTLSDaemon(t, ca, ca)

		info, err := load(t, "--backend-docker-host="+host, "--backend-docker-cert="+writeTestCerts(t, ca), "--backend-docker-tls-verify=true")
		require.NoError(t, err)
		assert.Equal(t, "linux/amd64", info.Platform)

		require.NotEmpty(t, daemon.userAgents)
		for _, ua := range daemon.userAgents {
			assert.Equal(t, userAgent, ua)
		}
		require.Len(t, daemon.clientCerts, len(daemon.userAgents), "every request has to present the client certificate")
		for _, cn := range daemon.clientCerts {
			assert.Equal(t, "woodpecker-agent", cn)
		}
	})

	t.Run("tls verify rejects unknown server certificate", func(t *testing.T) {
		ca := newTestCert(t, &x509.Certificate{Subject: pkix.Name{CommonName: "ca"}}, nil)
		otherCA := newTestCert(t, &x509.Certificate{Subject: pkix.Name{CommonName: "other-ca"}}, nil)
		daemon, host := newTLSDaemon(t, otherCA, ca)

		_, err := load(t, "--backend-docker-host="+host, "--backend-docker-cert="+writeTestCerts(t, ca), "--backend-docker-tls-verify=true")
		assert.ErrorContains(t, err, "certificate signed by unknown authority")
		assert.Empty(t, daemon.userAgents)
	})

	t.Run("tls without verify accepts unknown server certificate", func(t *testing.T) {
		ca := newTestCert(t, &x509.Certificate{Subject: pkix.Name{CommonName: "ca"}}, nil)
		otherCA := newTestCert(t, &x509.Certificate{Subject: pkix.Name{CommonName: "other-ca"}}, nil)
		daemon, host := newTLSDaemon(t, otherCA, ca)

		_, err := load(t, "--backend-docker-host="+host, "--backend-docker-cert="+writeTestCerts(t, ca), "--backend-docker-tls-verify=false")
		require.NoError(t, err)
		assert.NotEmpty(t, daemon.clientCerts)
	})

	t.Run("unusable certificates do not fall back to plain connection", func(t *testing.T) {
		daemon := &fakeDaemon{}
		srv := httptest.NewServer(daemon)
		t.Cleanup(srv.Close)

		_, err := load(t, "--backend-docker-host=tcp://"+srv.Listener.Addr().String(), "--backend-docker-cert="+t.TempDir())
		assert.ErrorContains(t, err, "ca.pem")
		assert.Empty(t, daemon.userAgents)
	})

	t.Run("plain custom host", func(t *testing.T) {
		daemon := &fakeDaemon{}
		srv := httptest.NewServer(daemon)
		t.Cleanup(srv.Close)

		info, err := load(t, "--backend-docker-host=tcp://"+srv.Listener.Addr().String())
		require.NoError(t, err)
		assert.Equal(t, "linux/amd64", info.Platform)

		require.NotEmpty(t, daemon.userAgents)
		for _, ua := range daemon.userAgents {
			assert.Equal(t, userAgent, ua)
		}
	})
}
