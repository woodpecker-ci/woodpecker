// Copyright 2022 Woodpecker Authors
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

import "github.com/woodpecker-ci/woodpecker/server/model"

type noEncryptionBuilder struct {
	clients []model.EncryptionClient
}

func (b noEncryptionBuilder) WithClients(clients []model.EncryptionClient) model.EncryptionServiceBuilder {
	b.clients = clients
	return b
}

func (b noEncryptionBuilder) Build() model.EncryptionService {
	svc := &noEncryption{}
	for _, client := range b.clients {
		client.SetEncryptionService(svc)
	}
	return svc
}

type noEncryption struct{}

func (svc *noEncryption) Encrypt(plaintext, _ string) string {
	return plaintext
}

func (svc *noEncryption) Decrypt(ciphertext, _ string) string {
	return ciphertext
}

func (svc *noEncryption) Disable() {}
