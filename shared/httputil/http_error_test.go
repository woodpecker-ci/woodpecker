// Copyright 2024 Woodpecker Authors
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

package httputil

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"os"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestEnhanceHTTPError checks that each recognized error class is mapped to its
// dedicated, human-readable prefix and that the original error stays wrapped.
func TestEnhanceHTTPError(t *testing.T) {
	t.Parallel()

	const endpoint = "https://example.com/api"

	tests := []struct {
		name       string
		err        error
		wantPrefix string
	}{
		{
			name: "context deadline exceeded",
			err:  context.DeadlineExceeded,
			// without the "Client.Timeout" marker it's a generic request timeout
			wantPrefix: "request timeout",
		},
		{
			name:       "client timeout deadline",
			err:        fmt.Errorf("wrapped: %w (Client.Timeout exceeded while awaiting headers)", context.DeadlineExceeded),
			wantPrefix: "connection timeout",
		},
		{
			name:       "context canceled",
			err:        context.Canceled,
			wantPrefix: "request canceled",
		},
		{
			name:       "DNS not found",
			err:        &net.DNSError{Err: "no such host", IsNotFound: true},
			wantPrefix: "DNS resolution failed",
		},
		{
			name:       "DNS timeout",
			err:        &net.DNSError{Err: "timeout", IsTimeout: true},
			wantPrefix: "DNS timeout",
		},
		{
			name:       "DNS generic",
			err:        &net.DNSError{Err: "server misbehaving"},
			wantPrefix: "DNS error",
		},
		{
			name:       "connection refused",
			err:        &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED},
			wantPrefix: "connection refused",
		},
		{
			name:       "connection reset",
			err:        &net.OpError{Op: "read", Err: syscall.ECONNRESET},
			wantPrefix: "connection reset",
		},
		{
			name:       "network unreachable",
			err:        &net.OpError{Op: "dial", Err: syscall.ENETUNREACH},
			wantPrefix: "network unreachable",
		},
		{
			name:       "host unreachable",
			err:        &net.OpError{Op: "dial", Err: syscall.EHOSTUNREACH},
			wantPrefix: "host unreachable",
		},
		{
			name:       "op error timeout",
			err:        &net.OpError{Op: "read", Err: timeoutErr{}},
			wantPrefix: "network timeout during read",
		},
		{
			name:       "op error generic",
			err:        &net.OpError{Op: "write", Err: errors.New("broken pipe")},
			wantPrefix: "network error during write",
		},
		{
			name:       "url error",
			err:        &url.Error{Op: "Get", URL: endpoint, Err: errors.New("malformed")},
			wantPrefix: "URL error",
		},
		{
			name:       "certificate invalid",
			err:        &x509.CertificateInvalidError{Cert: &x509.Certificate{}},
			wantPrefix: "TLS certificate invalid",
		},
		{
			name:       "unknown authority",
			err:        &x509.UnknownAuthorityError{},
			wantPrefix: "TLS certificate verification failed",
		},
		{
			name:       "hostname mismatch",
			err:        &x509.HostnameError{Certificate: &x509.Certificate{}, Host: "example.com"},
			wantPrefix: "TLS hostname mismatch",
		},
		{
			name:       "os invalid",
			err:        os.ErrInvalid,
			wantPrefix: "invalid argument",
		},
		{
			name:       "os permission",
			err:        os.ErrPermission,
			wantPrefix: "permission denied",
		},
		{
			name:       "os exist",
			err:        os.ErrExist,
			wantPrefix: "file already exists",
		},
		{
			name:       "os not exist",
			err:        os.ErrNotExist,
			wantPrefix: "file does not exist",
		},
		{
			name:       "os closed",
			err:        os.ErrClosed,
			wantPrefix: "file already closed",
		},
		{
			name:       "os no deadline",
			err:        os.ErrNoDeadline,
			wantPrefix: "file type does not support deadline",
		},
		{
			name:       "os deadline exceeded",
			err:        os.ErrDeadlineExceeded,
			wantPrefix: "i/o timeout",
		},
		{
			name:       "EOF",
			err:        io.EOF,
			wantPrefix: "unexpected connection closure",
		},
		{
			name:       "connection reset by peer string",
			err:        errors.New("read tcp: connection reset by peer"),
			wantPrefix: "connection reset by peer",
		},
		{
			name:       "generic error",
			err:        errors.New("some random error"),
			wantPrefix: "HTTP request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := EnhanceHTTPError(tt.err, "POST", endpoint)

			assert.Error(t, got)
			assert.Truef(t, strings.HasPrefix(got.Error(), tt.wantPrefix),
				"got %q, want prefix %q", got.Error(), tt.wantPrefix)
			// original error must remain unwrappable
			assert.ErrorIs(t, got, tt.err)
		})
	}
}

func TestEnhanceHTTPErrorNil(t *testing.T) {
	t.Parallel()
	assert.NoError(t, EnhanceHTTPError(nil, "POST", "https://example.com"))
}

func TestEnhanceHTTPErrorUnparsableEndpoint(t *testing.T) {
	t.Parallel()
	// a control character makes url.Parse fail, exercising the host fallback
	err := EnhanceHTTPError(io.EOF, "GET", "http://\x00bad")
	assert.Error(t, err)
	// fallback uses the raw endpoint as host
	assert.Contains(t, err.Error(), "\x00bad")
}

func TestEnhanceHTTPErrorPreservesOriginal(t *testing.T) {
	t.Parallel()
	enhanced := EnhanceHTTPError(io.EOF, "POST", "https://example.com/api")
	assert.ErrorIs(t, enhanced, io.EOF)
}

// timeoutErr is a minimal net.Error whose Timeout() reports true, used to drive
// the OpError timeout branch.
type timeoutErr struct{}

func (timeoutErr) Error() string   { return "i/o timeout" }
func (timeoutErr) Timeout() bool   { return true }
func (timeoutErr) Temporary() bool { return false }
