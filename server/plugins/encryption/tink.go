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
	"encoding/base64"
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/google/tink/go/tink"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

type tinkEncryptionService struct {
	keysetFilePath    string
	primaryKeyID      string
	encryption        tink.AEAD
	store             store.Store
	keysetFileWatcher *fsnotify.Watcher
	clients           []model.EncryptionClient
}

func (svc *tinkEncryptionService) Encrypt(plaintext, associatedData string) (string, error) {
	msg := []byte(plaintext)
	aad := []byte(associatedData)
	ciphertext, err := svc.encryption.Encrypt(msg, aad)
	if err != nil {
		return "", fmt.Errorf(errTemplateEncryptionFailed, err)
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (svc *tinkEncryptionService) Decrypt(ciphertext, associatedData string) (string, error) {
	ct, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf(errTemplateBase64DecryptionFailed, err)
	}

	plaintext, err := svc.encryption.Decrypt(ct, []byte(associatedData))
	if err != nil {
		return "", fmt.Errorf(errTemplateDecryptionFailed, err)
	}
	return string(plaintext), nil
}

func (svc *tinkEncryptionService) Disable() error {
	return svc.disable()
}
