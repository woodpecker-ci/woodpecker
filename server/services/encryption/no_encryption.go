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

import "go.woodpecker-ci.org/woodpecker/v2/server/services/encryption/types"

type noEncryptionBuilder struct {
	clients []types.EncryptionClient
}

func (b noEncryptionBuilder) WithClients(clients []types.EncryptionClient) types.EncryptionServiceBuilder {
	b.clients = clients
	return b
}

func (b noEncryptionBuilder) Build() (types.EncryptionService, error) {
	svc := &noEncryption{}
	for _, client := range b.clients {
		err := client.SetEncryptionService(svc)
		if err != nil {
			return nil, err
		}
	}
	return svc, nil
}

type noEncryption struct{}

func (svc *noEncryption) Encrypt(plaintext, _ string) (string, error) {
	return plaintext, nil
}

func (svc *noEncryption) Decrypt(ciphertext, _ string) (string, error) {
	return ciphertext, nil
}

func (svc *noEncryption) Disable() error {
	return nil
}
