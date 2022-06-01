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
	"crypto"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Verifier is used by by the HTTP server to verify the incoming HTTP requests
type Verifier struct {
	keyGetter       KeyGetter
	requiredHeaders []string
}

// NewVerifier creates a new Verifier using kg to get the key
// mapped to the ID received in the requests
func NewVerifier(kg KeyGetter) *Verifier {
	v := &Verifier{
		keyGetter: kg,
	}
	v.SetRequiredHeaders(nil)
	return v
}

// RequiredHeaders returns the required header the client have to include in
// the signature
func (v *Verifier) RequiredHeaders() []string {
	return append([]string{}, v.requiredHeaders...)
}

// SetRequiredHeaders set the list of headers to be included by the client to generate the signature
func (v *Verifier) SetRequiredHeaders(headers []string) {
	if len(headers) == 0 {
		headers = []string{"date"}
	}
	requiredHeaders := make([]string, 0, len(headers))
	for _, h := range headers {
		requiredHeaders = append(requiredHeaders, strings.ToLower(h))
	}
	v.requiredHeaders = requiredHeaders
}

// Verify parses req  and verify the signature using the key returned by
// the keyGetter. It returns the KeyId parameter from he signature header
// and a nil error if the signature verifies, an error otherwise
func (v *Verifier) Verify(req *http.Request) (string, error) {
	// retrieve and validate params from the request
	params := getParamsFromAuthHeader(req)
	if params == nil {
		return "", fmt.Errorf("no params present")
	}
	if params.KeyID == "" {
		return "", fmt.Errorf("keyId is required")
	}
	if params.Algorithm == "" {
		return "", fmt.Errorf("algorithm is required")
	}
	if len(params.Signature) == 0 {
		return "", fmt.Errorf("signature is required")
	}
	if len(params.Headers) == 0 {
		params.Headers = []string{"date"}
	}

header_check:
	for _, h := range v.requiredHeaders {
		for _, header := range params.Headers {
			if strings.EqualFold(h, header) {
				continue header_check
			}
		}
		return "", fmt.Errorf("missing required header in signature %q",
			h)
	}

	// calculate signature string for request
	sigData, err := BuildSignatureData(req, params.Headers, params.Created, params.Expires)
	if err != nil {
		return "", err
	}

	// look up key based on keyId
	key, err := v.keyGetter.GetKey(params.KeyID)
	if err != nil {
		return "", err
	}
	// we still leave this sanity check
	if key == nil {
		return "", fmt.Errorf("no key with id %q", params.KeyID)
	}

	switch params.Algorithm {
	case "rsa-sha1":
		rsaPubkey := toRSAPublicKey(key)
		if rsaPubkey == nil {
			return "", fmt.Errorf("algorithm %q is not supported by key %q",
				params.Algorithm, params.KeyID)
		}
		return params.KeyID, RSAVerify(rsaPubkey, crypto.SHA1, sigData, params.Signature)
	case "rsa-sha256":
		rsaPubkey := toRSAPublicKey(key)
		if rsaPubkey == nil {
			return "", fmt.Errorf("algorithm %q is not supported by key %q",
				params.Algorithm, params.KeyID)
		}
		return params.KeyID, RSAVerify(rsaPubkey, crypto.SHA256, sigData, params.Signature)
	case "hmac-sha256":
		hmacKey := toHMACKey(key)
		if hmacKey == nil {
			return "", fmt.Errorf("algorithm %q is not supported by key %q",
				params.Algorithm, params.KeyID)
		}
		return params.KeyID, HMACVerify(hmacKey, crypto.SHA256, sigData, params.Signature)
	case "ed25519":
		ed25519Key := toEd25519PublicKey(key)
		if ed25519Key == nil {
			return "", fmt.Errorf("algorithm %q is not supported by key %q",
				params.Algorithm, params.KeyID)
		}
		return params.KeyID, Ed25519Verify(ed25519Key, sigData, params.Signature)
	default:
		return "", fmt.Errorf("unsupported algorithm %q", params.Algorithm)
	}
}

// paramRE scans out recognized parameter keypairs. accepted values are those
// that are quoted
var paramRE = regexp.MustCompile(`(?U)\s*([a-zA-Z][a-zA-Z0-9_]*)\s*=\s*"(.*)"\s*`)

// Params holds the field requires to build the signature string
type Params struct {
	KeyID     string
	Algorithm string
	Headers   []string
	Signature []byte
	Created   time.Time
	Expires   time.Time
}

func getParamsFromAuthHeader(req *http.Request) *Params {
	return getParams(req, "Authorization", "Signature ")
}

func getParams(req *http.Request, header, prefix string) *Params {
	values := req.Header[http.CanonicalHeaderKey(header)]
	// last well-formed parameter wins
	for i := len(values) - 1; i >= 0; i-- {
		value := values[i]
		if prefix != "" {
			if trimmed := strings.TrimPrefix(value, prefix); trimmed != value {
				value = trimmed
			} else {
				continue
			}
		}

		matches := paramRE.FindAllStringSubmatch(value, -1)
		if matches == nil {
			continue
		}

		params := Params{}
		// malformed parameters get ignored.
		for _, match := range matches {
			switch match[1] {
			case "keyId":
				params.KeyID = match[2]
			case "algorithm":
				if algorithm, ok := parseAlgorithm(match[2]); ok {
					params.Algorithm = algorithm
				}
			case "headers":
				if headers, ok := parseHeaders(match[2]); ok {
					params.Headers = headers
				}
			case "signature":
				if signature, ok := parseSignature(match[2]); ok {
					params.Signature = signature
				}
			case "created":
				if created, ok := parseTime(match[2]); ok {
					params.Created = created
				}
			case "expires":
				if expires, ok := parseTime(match[2]); ok {
					params.Expires = expires
				}
			}
		}
		return &params
	}

	return nil
}

// parseAlgorithm parses recognized algorithm values
func parseAlgorithm(s string) (algorithm string, ok bool) {
	s = strings.TrimSpace(s)
	switch s {
	case "rsa-sha1", "rsa-sha256", "hmac-sha256", "ed25519":
		return s, true
	}
	return "", false
}

// parseHeaders parses a space separated list of header values.
func parseHeaders(s string) (headers []string, ok bool) {
	for _, header := range strings.Split(s, " ") {
		if header != "" {
			headers = append(headers, strings.ToLower(header))
		}
	}
	return headers, true
}

func parseSignature(s string) (signature []byte, ok bool) {
	signature, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, false
	}
	return signature, true
}

func parseTime(s string) (t time.Time, ok bool) {
	sec, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return t, false
	}
	return time.Unix(sec, 0), true
}
