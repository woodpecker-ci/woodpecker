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
	"errors"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/store/types"
	"golang.org/x/crypto/bcrypt"

	"golang.org/x/crypto/sha3"
)

func (svc *aesEncryptionService) loadCipher(password string) error {
	key, err := svc.hash([]byte(password))
	if err != nil {
		return fmt.Errorf(errTemplateAesFailedGeneratingKey, err)
	}
	keyHash, err := bcrypt.GenerateFromPassword(key, bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf(errTemplateAesFailedGeneratingKeyID, err)
	}
	svc.keyID = string(keyHash)

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
	if errors.Is(err, types.RecordNotExist) {
		return errEncryptionNotEnabled
	} else if err != nil {
		return fmt.Errorf(errTemplateFailedLoadingServerConfig, err)
	}

	plaintext, err := svc.Decrypt(ciphertextSample, keyIDAssociatedData)
	if plaintext != svc.keyID {
		return errEncryptionKeyInvalid
	} else if err != nil {
		return err
	}
	return nil
}

func (svc *aesEncryptionService) hash(data []byte) ([]byte, error) {
	result := make([]byte, 32)
	sha := sha3.NewShake256()

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
