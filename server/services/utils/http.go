// Copyright 2023 Woodpecker Authors
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

package utils

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ed25519"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/rs/zerolog/log"
	"github.com/yaronf/httpsign"

	host_matcher "go.woodpecker-ci.org/woodpecker/v3/server/services/utils/hostmatcher"
	"go.woodpecker-ci.org/woodpecker/v3/shared/httputil"
)

type Client struct {
	*httpsign.Client
}

func getHTTPClient(privateKey crypto.PrivateKey, allowedHostListValue string) (*httpsign.Client, error) {
	timeout := 10 * time.Second //nolint:mnd

	if allowedHostListValue == "" {
		allowedHostListValue = host_matcher.MatchBuiltinExternal
	}
	allowedHostMatcher := host_matcher.ParseHostMatchList("WOODPECKER_EXTENSIONS_ALLOWED_HOSTS", allowedHostListValue)

	pubKeyID := "woodpecker-ci-extensions"

	ed25519Key, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("invalid private key type")
	}

	signer, err := httpsign.NewEd25519Signer(ed25519Key,
		httpsign.NewSignConfig(),
		httpsign.Headers("@request-target", "content-digest")) // The Content-Digest header will be auto-generated
	if err != nil {
		return nil, err
	}

	// Create base transport with custom User-Agent
	baseTransport := httputil.NewUserAgentRoundTripper(
		&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
			DialContext:     host_matcher.NewDialContext("extensions", allowedHostMatcher),
		},
		"server-extensions",
	)

	client := http.Client{
		Timeout:   timeout,
		Transport: baseTransport,
	}

	config := httpsign.NewClientConfig().SetSignatureName(pubKeyID).SetSigner(signer)

	return httpsign.NewClient(client, config), nil
}

func NewHTTPClient(privateKey crypto.PrivateKey, allowedHostList string) (*Client, error) {
	client, err := getHTTPClient(privateKey, allowedHostList)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: client,
	}, nil
}

// Send makes an http request with retry logic.
func (e *Client) Send(ctx context.Context, method, path string, in, out any) (int, error) {
	// Maximum number of retries
	const maxRetries = 3
	// Initial backoff duration
	const initialBackoff = 500 * time.Millisecond
	// Maximum backoff interval
	const maxBackoffInterval = 5 * time.Second

	log.Debug().Msgf("HTTP request: %s %s, retries enabled (max: %d)", method, path, maxRetries)

	// Prepare request body bytes for possible retries
	var bodyBytes []byte
	if in != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(in); err != nil {
			return 0, err
		}
		bodyBytes = buf.Bytes()
	}

	// Parse URI once
	uri, err := url.Parse(path)
	if err != nil {
		return 0, err
	}

	// Retry loop with exponential backoff
	var statusCode int
	var lastErr error

	// Create backoff configuration
	backoffConfig := backoff.NewExponentialBackOff()
	backoffConfig.InitialInterval = initialBackoff
	backoffConfig.MaxInterval = maxBackoffInterval
	// No MaxElapsedTime, we'll handle max retries ourselves

	for retry := 0; retry <= maxRetries; retry++ {
		// Check if context is already canceled
		if ctx.Err() != nil {
			return 0, ctx.Err()
		}

		// Create request body for this attempt
		var body io.Reader
		if len(bodyBytes) > 0 {
			body = bytes.NewReader(bodyBytes)
		}

		// Create new request for each attempt
		req, err := http.NewRequestWithContext(ctx, method, uri.String(), body)
		if err != nil {
			return 0, err
		}
		if in != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		// Send request
		resp, err := e.Do(req)
		if err != nil {
			lastErr = err

			// Check if this is a retryable error
			if !isRetryableError(err) {
				log.Error().Err(err).Msgf("HTTP request failed (not retryable): %s %s", method, path)
				return 0, err
			}

			// If we've reached max retries, return the last error
			if retry == maxRetries {
				log.Error().Err(err).Msgf("HTTP request failed after %d retries: %s %s", maxRetries, method, path)
				return 0, err
			}

			// Wait with exponential backoff before retrying
			waitDuration := backoffConfig.NextBackOff()
			log.Debug().Err(err).Msgf("HTTP request failed, retrying in %v (attempt %d/%d): %s %s", waitDuration, retry+1, maxRetries, method, path)
			time.Sleep(waitDuration)
			continue
		}

		statusCode = resp.StatusCode
		// Read body immediately to ensure proper resource cleanup for retries
		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr

			// Check if this is a retryable error
			if isRetryableError(readErr) && retry < maxRetries {
				// Wait with exponential backoff before retrying
				waitDuration := backoffConfig.NextBackOff()
				log.Debug().Err(readErr).Msgf("HTTP response read failed, retrying in %v (attempt %d/%d): %s %s", waitDuration, retry+1, maxRetries, method, path)
				time.Sleep(waitDuration)
				continue
			}
			log.Error().Err(readErr).Msgf("HTTP response read failed (not retryable): %s %s", method, path)
			return statusCode, readErr
		}

		// Check if status code is retryable
		if isRetryableStatusCode(statusCode) {
			lastErr = fmt.Errorf("response: %d", statusCode)

			// If we've reached max retries, return the response body
			if retry == maxRetries {
				log.Error().Int("status", statusCode).Msgf("HTTP request failed after %d retries with status code: %s %s", maxRetries, method, path)
				return statusCode, fmt.Errorf("response: %s", string(respBody))
			}

			// Wait with exponential backoff before retrying
			waitDuration := backoffConfig.NextBackOff()
			log.Debug().Int("status", statusCode).Msgf("HTTP request returned retryable status code, retrying in %v (attempt %d/%d): %s %s", waitDuration, retry+1, maxRetries, method, path)
			time.Sleep(waitDuration)
			continue
		}

		// If status code is client error (4xx), don't retry
		if statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError {
			log.Debug().Int("status", statusCode).Msgf("HTTP request returned client error (not retryable): %s %s", method, path)
			return statusCode, fmt.Errorf("response: %s", string(respBody))
		}

		// If status code is OK (2xx), parse and return response
		if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
			if out != nil {
				err = json.NewDecoder(bytes.NewReader(respBody)).Decode(out)
				// Check for EOF error during response body parsing
				if err != nil && (errors.Is(err, io.EOF) || strings.Contains(err.Error(), "unexpected EOF")) {
					lastErr = err

					// If we've reached max retries, return the error
					if retry == maxRetries {
						log.Error().Err(err).Msgf("HTTP response parsing failed after %d retries: %s %s", maxRetries, method, path)
						return statusCode, err
					}

					// Wait with exponential backoff before retrying
					waitDuration := backoffConfig.NextBackOff()
					log.Debug().Err(err).Msgf("HTTP response parsing failed (EOF), retrying in %v (attempt %d/%d): %s %s", waitDuration, retry+1, maxRetries, method, path)
					time.Sleep(waitDuration)
					continue
				}
				if err != nil {
					log.Error().Err(err).Msgf("HTTP response parsing failed (not retryable): %s %s", method, path)
					return statusCode, err
				}
			}
			log.Debug().Int("status", statusCode).Msgf("HTTP request succeeded: %s %s", method, path)
			return statusCode, nil
		}

		// For any other status code, don't retry
		log.Error().Int("status", statusCode).Msgf("HTTP request returned unexpected status code (not retryable): %s %s", method, path)
		return statusCode, fmt.Errorf("response: %s", string(respBody))
	}

	return statusCode, lastErr
}

// isRetryableError checks if an error is transient and suitable for retry.
func isRetryableError(err error) bool {
	// Check for network-related errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Retry on timeout errors
		if netErr.Timeout() {
			return true
		}
	}

	// Check for specific error types
	switch {
	case errors.Is(err, net.ErrClosed),
		errors.Is(err, io.EOF),
		errors.Is(err, io.ErrUnexpectedEOF):
		return true
	}

	// Check for error strings that indicate retryable conditions
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset by peer") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "TLS handshake timeout")
}

// isRetryableStatusCode checks if an HTTP status code is suitable for retry.
func isRetryableStatusCode(statusCode int) bool {
	// Retry on server errors (5xx)
	return statusCode >= http.StatusInternalServerError && statusCode < http.StatusNetworkAuthenticationRequired
}
