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
