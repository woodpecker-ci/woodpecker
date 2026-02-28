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
	"io"
	"net"
	"syscall"
	"testing"
)

// TestEnhanceHTTPError tests the enhanceHTTPError function with various error types.
func TestEnhanceHTTPError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		method   string
		endpoint string
		want     string
	}{
		{
			name:     "nil error",
			err:      nil,
			method:   "POST",
			endpoint: "https://example.com",
			want:     "",
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			method:   "POST",
			endpoint: "https://example.com/api",
			want:     "request timeout",
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			method:   "GET",
			endpoint: "https://example.com/api",
			want:     "request canceled",
		},
		{
			name: "DNS not found error",
			err: &net.DNSError{
				Err:         "no such host",
				IsNotFound:  true,
				IsTimeout:   false,
				IsTemporary: false,
			},
			method:   "POST",
			endpoint: "https://nonexistent.example.com",
			want:     "DNS resolution failed",
		},
		{
			name: "DNS timeout error",
			err: &net.DNSError{
				Err:         "timeout",
				IsTimeout:   true,
				IsTemporary: true,
			},
			method:   "POST",
			endpoint: "https://example.com",
			want:     "DNS timeout",
		},
		{
			name:     "unknown authority certificate error",
			err:      x509.UnknownAuthorityError{},
			method:   "POST",
			endpoint: "https://self-signed.example.com",
			want:     "TLS certificate verification failed",
		},
		{
			name: "connection refused",
			err: &net.OpError{
				Op:  "dial",
				Err: syscall.ECONNREFUSED,
			},
			method:   "POST",
			endpoint: "https://localhost:9999",
			want:     "connection refused",
		},
		{
			name: "connection reset",
			err: &net.OpError{
				Op:  "read",
				Err: syscall.ECONNRESET,
			},
			method:   "POST",
			endpoint: "https://example.com",
			want:     "connection reset",
		},
		{
			name:     "EOF error",
			err:      io.EOF,
			method:   "POST",
			endpoint: "https://example.com/api",
			want:     "unexpected connection closure",
		},
		{
			name:     "generic error",
			err:      errors.New("some random error"),
			method:   "POST",
			endpoint: "https://example.com",
			want:     "HTTP request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EnhanceHTTPError(tt.err, tt.method, tt.endpoint)

			if tt.want == "" {
				if got != nil {
					t.Errorf("enhanceHTTPError() = %v, want nil", got)
				}
				return
			}

			if got == nil {
				t.Errorf("enhanceHTTPError() = nil, want error containing %q", tt.want)
				return
			}

			if got.Error() == "" {
				t.Errorf("enhanceHTTPError() returned empty error message")
				return
			}

			// check empty error message
			errMsg := got.Error()
			if len(errMsg) == 0 {
				t.Errorf("enhanceHTTPError() returned empty error string")
				return
			}

			// view the enhance error message
			t.Logf("enhanced error: %v", errMsg)
		})
	}
}

func TestEnhanceHTTPErrorPreservesOriginal(t *testing.T) {
	originalErr := io.EOF
	endpoint := "https://example.com/api"

	enhanced := EnhanceHTTPError(originalErr, "POST", endpoint)

	// the io.EOF error should be wrapped inside the enhanced error
	if !errors.Is(enhanced, originalErr) {
		t.Errorf("enhanced error should wrap original error, but errors.Is returned false")
	}
}
