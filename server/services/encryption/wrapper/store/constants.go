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

package store

const (
	errMessageTemplateFailedToEnable        = "failed enabling secret store encryption: %w"
	errMessageTemplateFailedToMigrate       = "failed migrating secret store encryption: %w"
	errMessageTemplateFailedToEncryptSecret = "failed to encrypt secret id=%d: %w"
	errMessageTemplateFailedToDecryptSecret = "failed to decrypt secret id=%d: %w"
	errMessageTemplateStorageError          = "Storage error: could not update secret in DB"

	errMessageTemplateFailedToRollbackSecretCreation = "failed creating secret: %w. Also failed deleting temporary secret record from store: %s"

	errMessageInitSeveralTimes = "attempt to init encrypted storage more than once"

	logMessageEnablingSecretsEncryption         = "Encrypting all secrets in database"
	logMessageEnablingSecretsEncryptionSuccess  = "All secrets are encrypted"
	logMessageMigratingSecretsEncryption        = "Migrating encryption keys"
	logMessageMigratingSecretsEncryptionSuccess = "Secrets encryption migrated successfully"
)

// Registry wrapper.
const (
	errMessageTemplateFailedToEnableRegistries         = "failed enabling registry store encryption: %w"
	errMessageTemplateFailedToMigrateRegistries        = "failed migrating registry store encryption: %w"
	errMessageTemplateFailedToEncryptRegistry          = "failed to encrypt registry id=%d: %w"
	errMessageTemplateFailedToDecryptRegistry          = "failed to decrypt registry id=%d: %w"
	errMessageTemplateRegistryStorageError             = "Storage error: could not update registry in DB"
	errMessageTemplateFailedToRollbackRegistryCreation = "failed creating registry: %w. Also failed deleting temporary registry record from store: %s"

	logMessageEnablingRegistriesEncryption         = "Encrypting all registry passwords in database"
	logMessageEnablingRegistriesEncryptionSuccess  = "All registry passwords are encrypted"
	logMessageMigratingRegistriesEncryption        = "Migrating registry password encryption keys"
	logMessageMigratingRegistriesEncryptionSuccess = "Registry password encryption migrated successfully"
)

// User wrapper.
const (
	errMessageTemplateFailedToEnableUsers  = "failed enabling user token encryption: %w"
	errMessageTemplateFailedToMigrateUsers = "failed migrating user token encryption: %w"
	errMessageTemplateFailedToEncryptUser  = "failed to encrypt tokens of user id=%d: %w"
	errMessageTemplateFailedToDecryptUser  = "failed to decrypt tokens of user id=%d: %w"
	errMessageTemplateUserStorageError     = "Storage error: could not update user in DB"

	logMessageEnablingUsersEncryption         = "Encrypting all user tokens in database"
	logMessageEnablingUsersEncryptionSuccess  = "All user tokens are encrypted"
	logMessageMigratingUsersEncryption        = "Migrating user token encryption keys"
	logMessageMigratingUsersEncryptionSuccess = "User token encryption migrated successfully"
)
