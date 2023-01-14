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

package encrypted

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
