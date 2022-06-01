// Copyright (C) 2017 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpsig

import (
	ed "crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	// Rand is a hookable reader used as a random byte source.
	Rand io.Reader = rand.Reader
)

// requestPath returns the :path pseudo header according to the HTTP/2 spec.
func requestPath(req *http.Request) string {
	path := req.URL.Path
	if path == "" {
		path = "/"
	}
	if req.URL.RawQuery != "" {
		path += "?" + req.URL.RawQuery
	}
	return path
}

// BuildSignatureString constructs a signature string following section 2.3
func BuildSignatureString(req *http.Request, headers []string, created, expires time.Time) (string, error) {
	if len(headers) == 0 {
		headers = []string{"(created)"}
	}

	values := make([]string, 0, len(headers))
	for _, h := range headers {

		switch h {
		case "(request-target)":
			values = append(values, fmt.Sprintf("%s: %s %s",
				h, strings.ToLower(req.Method), requestPath(req)))
		case "(created)":
			values = append(values, fmt.Sprintf("%s: %d", h, created.Unix()))
		case "(expires)":
			values = append(values, fmt.Sprintf("%s: %d", h, expires.Unix()))
		case "host":
			values = append(values, fmt.Sprintf("%s: %s", h, req.Host))
		case "date":
			if req.Header.Get(h) == "" {
				req.Header.Set(h, time.Now().UTC().Format(http.TimeFormat))
			}
			values = append(values, fmt.Sprintf("%s: %s", h, req.Header.Get(h)))
		default:
			vs, found := req.Header[http.CanonicalHeaderKey(h)]
			if !found {
				return "", fmt.Errorf("expected %s to exists", h)
			}
			for _, v := range vs {
				values = append(values, fmt.Sprintf("%s: %s", h, strings.TrimSpace(v)))
			}
		}
	}
	return strings.Join(values, "\n"), nil
}

// BuildSignatureData is a convenience wrapper around BuildSignatureString that
// returns []byte instead of a string.
func BuildSignatureData(req *http.Request, headers []string, created, expires time.Time) ([]byte, error) {
	s, err := BuildSignatureString(req, headers, created, expires)
	return []byte(s), err
}

func toRSAPrivateKey(key interface{}) *rsa.PrivateKey {
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return k
	default:
		return nil
	}
}

func toRSAPublicKey(key interface{}) *rsa.PublicKey {
	switch k := key.(type) {
	case *rsa.PublicKey:
		return k
	case *rsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

func toHMACKey(key interface{}) []byte {
	switch k := key.(type) {
	case []byte:
		return k
	default:
		return nil
	}
}

func toEd25519PrivateKey(key interface{}) ed.PrivateKey {
	switch k := key.(type) {
	case ed.PrivateKey:
		return k
	default:
		return nil
	}
}

func toEd25519PublicKey(key interface{}) ed.PublicKey {
	switch k := key.(type) {
	case ed.PublicKey:
		return k
	case ed.PrivateKey:
		return k.Public().(ed.PublicKey)
	default:
		return nil
	}
}

func unsupportedAlgorithm(a Algorithm) error {
	return fmt.Errorf("key does not support algorithm %q", a.Name())
}
