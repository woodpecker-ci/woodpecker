// Copyright 2025 Woodpecker Authors
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
	"net"
	"net/url"
	"os"
	"strings"
	"syscall"
)

// EnhanceHTTPError adds detailed context to HTTP errors to help with debugging.
func EnhanceHTTPError(err error, method, endpoint string) error {
	if err == nil {
		return nil
	}

	// parse url to get host information
	parsedURL, parseErr := url.Parse(endpoint)
	var host string
	if parseErr == nil {
		host = parsedURL.Host
	} else {
		host = endpoint
	}

	// base error message
	baseMsg := fmt.Sprintf("%s %q", method, endpoint)

	// check for context errors
	// timeout
	if errors.Is(err, context.DeadlineExceeded) {
		if strings.Contains(err.Error(), "Client.Timeout") {
			return fmt.Errorf("connection timeout: %s: %w (the remote server at %s did not respond within the configured timeout)", baseMsg, err, host)
		}
		return fmt.Errorf("request timeout: %s: %w (operation took too long time)", baseMsg, err)
	}

	// cancellation
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("request canceled: %s: %w (the operation was canceled before completion)", baseMsg, err)
	}

	// check for net package errors
	// dns error handling
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		if dnsErr.IsNotFound {
			return fmt.Errorf("DNS resolution failed: %s: %w (hostname %s does not exist or cannot be resolved)", baseMsg, err, host)
		}
		if dnsErr.IsTimeout {
			return fmt.Errorf("DNS timeout: %s: %w (DNS server did not respond in time)", baseMsg, err)
		}
		return fmt.Errorf("DNS error: %s: %w", baseMsg, err)
	}

	// op error handling
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		// connection refused
		if errors.Is(opErr.Err, syscall.ECONNREFUSED) {
			return fmt.Errorf("connection refused: %s: %w (server at %s is not accepting connections - is it running?)", baseMsg, err, host)
		}

		// connection reset
		if errors.Is(opErr.Err, syscall.ECONNRESET) {
			return fmt.Errorf("connection reset: %s: %w (server at %s closed the connection unexpectedly)", baseMsg, err, host)
		}

		// network unreachable
		if errors.Is(opErr.Err, syscall.ENETUNREACH) {
			return fmt.Errorf("network unreachable: %s: %w (cannot reach %s - check network connectivity)", baseMsg, err, host)
		}

		// host unreachable
		if errors.Is(opErr.Err, syscall.EHOSTUNREACH) {
			return fmt.Errorf("host unreachable: %s: %w (cannot reach %s - host may be down or firewall blocking)", baseMsg, err, host)
		}

		// timeout during operation
		if opErr.Timeout() {
			return fmt.Errorf("network timeout during %s: %s: %w (operation at %s took too long)", opErr.Op, baseMsg, err, host)
		}

		return fmt.Errorf("network error during %s: %s: %w", opErr.Op, baseMsg, err)
	}

	// check for url parsing errors
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return fmt.Errorf("URL error: %s: %w (check if the endpoint URL is correctly formatted)", baseMsg, err)
	}

	// check for TLS/certificate errors
	var certErr *x509.CertificateInvalidError
	if errors.As(err, &certErr) {
		return fmt.Errorf("TLS certificate invalid: %s: %w (certificate validation failed for %s)", baseMsg, err, host)
	}

	var unknownAuthErr *x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthErr) {
		return fmt.Errorf("TLS certificate verification failed: %s: %w (certificate signed by unknown authority for %s)", baseMsg, err, host)
	}

	var hostErr *x509.HostnameError
	if errors.As(err, &hostErr) {
		return fmt.Errorf("TLS hostname mismatch: %s: %w (certificate is not valid for %s)", baseMsg, err, host)
	}

	// check for os errors
	if errors.Is(err, os.ErrInvalid) {
		return fmt.Errorf("invalid argument: %s: %w", baseMsg, err)
	}
	if errors.Is(err, os.ErrPermission) {
		return fmt.Errorf("permission denied: %s: %w", baseMsg, err)
	}
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("file already exists: %s: %w", baseMsg, err)
	}
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file does not exist: %s: %w", baseMsg, err)
	}
	if errors.Is(err, os.ErrClosed) {
		return fmt.Errorf("file already closed: %s: %w", baseMsg, err)
	}
	if errors.Is(err, os.ErrNoDeadline) {
		return fmt.Errorf("file type does not support deadline: %s: %w", baseMsg, err)
	}
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return fmt.Errorf("i/o timeout: %s: %w", baseMsg, err)
	}

	// check for EOF specifically
	if err.Error() == "EOF" || strings.Contains(err.Error(), "EOF") {
		return fmt.Errorf("unexpected connection closure: %s: %w (server at %s closed connection prematurely - possible causes: server crash, request too large, incompatible protocol, or server-side timeout)", baseMsg, err, host)
	}

	// check for "connection reset by peer"
	if strings.Contains(err.Error(), "connection reset by peer") {
		return fmt.Errorf("connection reset by peer: %s: %w (server at %s forcibly closed the connection)", baseMsg, err, host)
	}

	// generic error
	return fmt.Errorf("HTTP request failed: %s: %w", baseMsg, err)
}
