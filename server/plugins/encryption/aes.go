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
	"encoding/base64"
	"fmt"

	"github.com/google/tink/go/subtle/random"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/sha3"
)

const (
	AesAlgo            = "aes"
	Sha256Size         = 32
	AESGCMSIVNonceSize = 12
)

type aesEncryptionService struct {
	keyId  string
	cipher cipher.AEAD
}

func NewAes(password string) (EncryptionService, error) {
	log.Debug().Msg("initializing AES encryption service")

	key, err := hash([]byte(password))
	if err != nil {
		return nil, NewKeyGenerationError(err)
	}

	keyHash, err := bcrypt.GenerateFromPassword(key, bcrypt.DefaultCost)
	if err != nil {
		return nil, NewKeyGenerationIdError(err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, NewCipherLoadingError(err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, NewCipherLoadingError(err)
	}

	service := aesEncryptionService{
		keyId:  string(keyHash),
		cipher: aead,
	}

	log.Debug().Msg("AES encryption service has been initialized")
	return &service, nil
}

func hash(data []byte) ([]byte, error) {
	result := make([]byte, Sha256Size)
	sha := sha3.NewShake256()

	_, err := sha.Write(data)
	if err != nil {
		return nil, NewHashCalculationError(err)
	}
	_, err = sha.Read(result)
	if err != nil {
		return nil, NewHashCalculationError(err)
	}
	return result, nil
}

func (svc *aesEncryptionService) Algo() string {
	return AesAlgo
}

func (svc *aesEncryptionService) Encrypt(plaintext, associatedData string) (string, error) {
	msg := []byte(plaintext)
	aad := []byte(associatedData)

	nonce := random.GetRandomBytes(uint32(AESGCMSIVNonceSize))
	ciphertext := svc.cipher.Seal(nil, nonce, msg, aad)

	result := make([]byte, 0, AESGCMSIVNonceSize+len(ciphertext))
	result = append(result, nonce...)
	result = append(result, ciphertext...)

	return base64.RawStdEncoding.EncodeToString(result), nil
}

func (svc *aesEncryptionService) Decrypt(ciphertext, associatedData string) (string, error) {
	bytes, err := base64.RawStdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", NewBase64DecryptionError(err)
	}

	nonce := bytes[:AESGCMSIVNonceSize]
	message := bytes[AESGCMSIVNonceSize:]

	plaintext, err := svc.cipher.Open(nil, nonce, message, []byte(associatedData))
	if err != nil {
		return "", NewDecryptionError(err)
	}
	return string(plaintext), nil
}

func NewHashCalculationError(e error) error {
	return fmt.Errorf("failed calculating hash: %w", e)
}

func NewKeyGenerationError(e error) error {
	return fmt.Errorf("failed generating key from passphrase: %w", e)
}

func NewKeyGenerationIdError(e error) error {
	return fmt.Errorf("failed generating key id: %w", e)
}

func NewCipherLoadingError(e error) error {
	return fmt.Errorf("failed loading encryption cipher: %w", e)
}

func NewBase64DecryptionError(e error) error {
	return fmt.Errorf("Base64 decryption failed: %w", e)
}

func NewDecryptionError(e error) error {
	return fmt.Errorf("decryption error: %w", e)
}
