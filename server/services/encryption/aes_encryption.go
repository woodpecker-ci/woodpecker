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
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha3"
	"encoding/hex"
	"errors"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v3/server/store/types"
)

func (svc *aesEncryptionService) loadCipher(password string) error {
	key := sha3.SumSHAKE256([]byte(password), aesKeySize)
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf(errTemplateAesFailedLoadingCipher, err)
	}
	aead, err := cipher.NewGCMWithRandomNonce(block)
	if err != nil {
		return fmt.Errorf(errTemplateAesFailedLoadingCipher, err)
	}
	svc.cipher = aead

	// The key id is derived with domain separation, so storing it (encrypted)
	// leaks nothing about the key. It must be deterministic: it is compared
	// against the sample stored in the database to validate the key across
	// server restarts.
	keyID := sha3.SumSHAKE256(append([]byte(keyIDDomainSeparation), key...), aesKeySize)
	svc.keyID = hex.EncodeToString(keyID)
	return nil
}

func (svc *aesEncryptionService) validateKey() error {
	ciphertextSample, err := svc.store.ServerConfigGet(ciphertextSampleConfigKey)
	if errors.Is(err, types.ErrRecordNotExist) {
		return errEncryptionNotEnabled
	} else if err != nil {
		return fmt.Errorf(errTemplateFailedLoadingServerConfig, err)
	}

	plaintext, err := svc.Decrypt(ciphertextSample, keyIDAssociatedData)
	if err != nil {
		return errEncryptionKeyInvalid
	}
	if plaintext != svc.keyID {
		return errEncryptionKeyInvalid
	}
	return nil
}
