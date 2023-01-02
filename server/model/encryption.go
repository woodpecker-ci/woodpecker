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

package model

// EncryptionBuilder is user API to obtain correctly configured encryption
type EncryptionBuilder interface {
	WithClient(client EncryptionClient) EncryptionBuilder
	Build() error
}

type EncryptionServiceBuilder interface {
	WithClients(clients []EncryptionClient) EncryptionServiceBuilder
	Build() (EncryptionService, error)
}

type EncryptionService interface {
	Encrypt(plaintext, associatedData string) (string, error)
	Decrypt(ciphertext, associatedData string) (string, error)
	Disable() error
}

type EncryptionClient interface {
	// SetEncryptionService should be used only by EncryptionServiceBuilder
	SetEncryptionService(encryption EncryptionService) error
	// EnableEncryption should encrypt all service data
	EnableEncryption() error
	// MigrateEncryption should decrypt all existing data and encrypt it with new encryption service
	MigrateEncryption(newEncryption EncryptionService) error
}
