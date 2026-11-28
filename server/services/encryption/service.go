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

	"github.com/tink-crypto/tink-go/v2/tink"
)

// encrypt encrypts value with key, bound to the associated data, and encodes
// the ciphertext as base64.
func encrypt(key tink.AEAD, value, associatedData string) (string, error) {
	ciphertext, err := key.Encrypt([]byte(value), []byte(associatedData))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decrypt reverses encrypt.
func decrypt(key tink.AEAD, value, associatedData string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	plaintext, err := key.Decrypt(ciphertext, []byte(associatedData))
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
