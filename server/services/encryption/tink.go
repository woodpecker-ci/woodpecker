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
	"github.com/tink-crypto/tink-go/v2/tink"

	"go.woodpecker-ci.org/woodpecker/v3/server/services/encryption/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

type tinkEncryptionService struct {
	keysetFilePath string
	primaryKeyID   string
	encryption     tink.AEAD
	store          store.Store
	clients        []types.EncryptionClient
}

func (svc *tinkEncryptionService) Encrypt(plaintext, associatedData string) (string, error) {
	return encrypt(svc.encryption, plaintext, associatedData)
}

func (svc *tinkEncryptionService) Decrypt(ciphertext, associatedData string) (string, error) {
	return decrypt(svc.encryption, ciphertext, associatedData)
}

func (svc *tinkEncryptionService) Disable() error {
	return svc.disable()
}
