package model

// EncryptionBuilder is user API to obtain correctly configured encryption
type EncryptionBuilder interface {
	WithClient(client EncryptionClient) EncryptionBuilder
	Build()
}

// EncryptionServiceBuilder should be used only in encryption configuration process
type EncryptionServiceBuilder interface {
	WithClients(clients []EncryptionClient) EncryptionServiceBuilder
	Build() EncryptionService
}

// EncryptionService defines a service for encryption.
type EncryptionService interface {
	Encrypt(plaintext string, associatedData string) string
	Decrypt(ciphertext string, associatedData string) string
	Disable()
}

// EncryptionClient should be used only in encryption configuration process
type EncryptionClient interface {
	// InitEncryption should be available only once
	InitEncryption(encryption EncryptionService)
	// EnableEncryption should encrypt all service data
	EnableEncryption()
	// MigrateEncryption should decrypt all existing data and encrypt it with new encryption service
	MigrateEncryption(newEncryption EncryptionService)
}
