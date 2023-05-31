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

import "errors"

// common
const (
	rawKeyConfigFlag             = "encryption-raw-key"
	tinkKeysetFilepathConfigFlag = "encryption-tink-keyset"
	disableEncryptionConfigFlag  = "encryption-disable-flag"

	ciphertextSampleConfigKey = "encryption-ciphertext-sample"

	keyTypeTink = "tink"
	keyTypeRaw  = "raw"
	keyTypeNone = "none"

	keyIDAssociatedData = "Primary key id"
	AESGCMSIVNonceSize  = 12
)

var (
	errEncryptionNotEnabled = errors.New("encryption is not enabled")
	errEncryptionKeyInvalid = errors.New("encryption key is invalid")
	errEncryptionKeyRotated = errors.New("encryption key is being rotated")
)

const (
	// error wrapping templates
	errTemplateFailedInitializingUnencrypted = "failed initializing server in unencrypted mode: %w"
	errTemplateFailedInitializing            = "failed initializing encryption service: %w"
	errTemplateFailedEnablingEncryption      = "failed enabling encryption: %w"
	errTemplateFailedRotatingEncryption      = "failed rotating encryption: %w"
	errTemplateFailedDisablingEncryption     = "failed disabling encryption: %w"
	errTemplateFailedLoadingServerConfig     = "failed to load server encryption config: %w"
	errTemplateFailedUpdatingServerConfig    = "failed updating server encryption configuration: %w"
	errTemplateFailedInitializingClients     = "failed initializing encryption clients: %w"
	errTemplateFailedValidatingKey           = "failed validating encryption key: %w"
	errTemplateEncryptionFailed              = "encryption error: %w"
	errTemplateBase64DecryptionFailed        = "decryption error: Base64 decryption failed. Cause: %w"
	errTemplateDecryptionFailed              = "decryption error: %w"

	// error messages
	errMessageTemplateUnsupportedKeyType = "unsupported encryption key type: %s"
	errMessageCantUseBothServices        = "can not use raw encryption key and tink keyset at the same time"
	errMessageNoKeysProvided             = "encryption enabled but no keys provided"
	errMessageFailedRotatingEncryption   = "failed rotating encryption"

	// log messages
	logMessageEncryptionEnabled       = "encryption enabled"
	logMessageEncryptionDisabled      = "encryption disabled"
	logMessageEncryptionKeyRegistered = "registered new encryption key"
	logMessageClientsInitialized      = "initialized encryption on registered clients"
	logMessageClientsEnabled          = "enabled encryption on registered services"
	logMessageClientsRotated          = "updated encryption key on registered services"
	logMessageClientsDecrypted        = "disabled encryption on registered services"
)

// tink
const (
	// error wrapping templates
	errTemplateTinkFailedLoadingKeyset              = "failed loading encryption keyset: %w"
	errTemplateTinkFailedValidatingKeyset           = "failed validating encryption keyset: %w"
	errTemplateTinkFailedInitializeFileWatcher      = "failed initializing keyset file watcher: %w"
	errTemplateTinkFailedSubscribeKeysetFileChanges = "failed subscribing on encryption keyset file changes: %w"
	errTemplateTinkFailedOpeningKeyset              = "failed opening encryption keyset file: %w"
	errTemplateTinkFailedReadingKeyset              = "failed reading encryption keyset from file: %w"
	errTemplateTinkFailedInitializingAEAD           = "failed initializing AEAD instance: %w"

	// error messages
	errMessageTinkKeysetFileWatchFailed = "failed watching encryption keyset file changes"

	// log message templates
	logTemplateTinkKeysetFileChanged       = "changes detected in encryption keyset file: '%s'. Encryption service will be reloaded"
	logTemplateTinkLoadingKeyset           = "loading encryption keyset from file: %s"
	logTemplateTinkFailedClosingKeysetFile = "could not close keyset file: %s"
)

// aes
const (
	// error wrapping templates
	errTemplateAesFailedLoadingCipher   = "failed loading encryption cipher: %w"
	errTemplateAesFailedCalculatingHash = "failed calculating hash: %w"
	errTemplateAesFailedGeneratingKey   = "failed generating key from passphrase: %w"
	errTemplateAesFailedGeneratingKeyID = "failed generating key id: %w"
)
