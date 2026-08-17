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
	key, err := svc.hash([]byte(password))
	if err != nil {
		return fmt.Errorf(errTemplateAesFailedGeneratingKey, err)
	}
	keyID, err := svc.deriveKeyID(key)
	if err != nil {
		return fmt.Errorf(errTemplateAesFailedGeneratingKeyID, err)
	}
	svc.keyID = keyID

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf(errTemplateAesFailedLoadingCipher, err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf(errTemplateAesFailedLoadingCipher, err)
	}
	svc.cipher = aead
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

// deriveKeyID derives a deterministic, non-reversible identifier for the
// encryption key. Domain separation ensures the id differs from the key
// derivation of the password, so storing it (encrypted) leaks nothing about
// the key itself. It must be deterministic: it is compared against the
// sample stored in the database to validate the key across server restarts.
func (svc *aesEncryptionService) deriveKeyID(key []byte) (string, error) {
	result := make([]byte, 32)
	sha := sha3.NewSHAKE256()

	_, err := sha.Write([]byte(keyIDDomainSeparation))
	if err != nil {
		return "", fmt.Errorf(errTemplateAesFailedCalculatingHash, err)
	}
	_, err = sha.Write(key)
	if err != nil {
		return "", fmt.Errorf(errTemplateAesFailedCalculatingHash, err)
	}
	_, err = sha.Read(result)
	if err != nil {
		return "", fmt.Errorf(errTemplateAesFailedCalculatingHash, err)
	}
	return hex.EncodeToString(result), nil
}

func (svc *aesEncryptionService) hash(data []byte) ([]byte, error) {
	result := make([]byte, 32)
	sha := sha3.NewSHAKE256()

	_, err := sha.Write(data)
	if err != nil {
		return nil, fmt.Errorf(errTemplateAesFailedCalculatingHash, err)
	}
	_, err = sha.Read(result)
	if err != nil {
		return nil, fmt.Errorf(errTemplateAesFailedCalculatingHash, err)
	}
	return result, nil
}
