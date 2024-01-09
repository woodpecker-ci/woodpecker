// Copyright 2023 Woodpecker Authors
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
	"crypto/cipher"
	"encoding/base64"
	"fmt"

	"github.com/google/tink/go/subtle/random"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
)

type aesEncryptionService struct {
	cipher  cipher.AEAD
	keyID   string
	store   store.Store
	clients []model.EncryptionClient
}

func (svc *aesEncryptionService) Encrypt(plaintext, associatedData string) (string, error) {
	msg := []byte(plaintext)
	aad := []byte(associatedData)

	nonce := random.GetRandomBytes(uint32(AESGCMSIVNonceSize))
	ciphertext := svc.cipher.Seal(nil, nonce, msg, aad)

	result := make([]byte, 0, AESGCMSIVNonceSize+len(ciphertext))
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return base64.StdEncoding.EncodeToString(result), nil
}

func (svc *aesEncryptionService) Decrypt(ciphertext, associatedData string) (string, error) {
	bytes, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf(errTemplateBase64DecryptionFailed, err)
	}

	nonce := bytes[:AESGCMSIVNonceSize]
	message := bytes[AESGCMSIVNonceSize:]

	plaintext, err := svc.cipher.Open(nil, nonce, message, []byte(associatedData))
	if err != nil {
		return "", fmt.Errorf(errTemplateDecryptionFailed, err)
	}
	return string(plaintext), nil
}

func (svc *aesEncryptionService) Disable() error {
	return svc.disable()
}
