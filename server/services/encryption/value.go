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

package encryption

import (
	"encoding/base64"
	"errors"
	"strings"

	"github.com/tink-crypto/tink-go/v2/tink"
)

// encryptedPrefix marks stored values that hold ciphertext. Values without it
// are plaintext, so data stays readable while it is being converted.
const encryptedPrefix = "woodpecker-enc:v1:"

var errNoKey = errors.New("value is encrypted, but no encryption key is configured")

// encrypt encrypts value with key, bound to the associated data. A nil key
// keeps the value in plaintext.
func encrypt(key tink.AEAD, value, associatedData string) (string, error) {
	if key == nil {
		return value, nil
	}
	ciphertext, err := key.Encrypt([]byte(value), []byte(associatedData))
	if err != nil {
		return "", err
	}
	return encryptedPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt reverses encrypt. Plaintext values are returned as they are.
func decrypt(key tink.AEAD, value, associatedData string) (string, error) {
	encoded, encrypted := strings.CutPrefix(value, encryptedPrefix)
	if !encrypted {
		return value, nil
	}
	if key == nil {
		return "", errNoKey
	}
	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	plaintext, err := key.Decrypt(ciphertext, []byte(associatedData))
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// convert re-encrypts value from key from to key to, where a nil key means
// plaintext. Unless all is set, a value that already has the form of key to
// (plaintext or encrypted) is returned unchanged.
func convert(value, associatedData string, from, to tink.AEAD, all bool) (string, error) {
	if !all && strings.HasPrefix(value, encryptedPrefix) == (to != nil) {
		return value, nil
	}
	plaintext, err := decrypt(from, value, associatedData)
	if err != nil {
		return "", err
	}
	return encrypt(to, plaintext, associatedData)
}
