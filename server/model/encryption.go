package model

const (
	TinkEncryptionType            = "tink"
	SimpleSymmetricEncryptionType = "simple_symmetric"
	DisabledEncryptionType        = "none"
)

// EncryptionBuilder is user API to obtain correctly configured encryption
type EncryptionBuilder interface {
	OfType(encryptionType string) EncryptionBuilder
	WithClient(client EncryptionClient) EncryptionBuilder
	Init() EncryptionService
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
	InitEncryption(encryption EncryptionService)
	// OnEnableEncryption should encrypt all service data
	OnEnableEncryption()
	// OnRotateEncryption should decrypt all existing data and encrypt it with new encryption
	OnRotateEncryption(newEncryption EncryptionService)
	// OnDisableEncryption should decrypt all data and guarantee that EncryptionClient service will stop processing requests
	OnDisableEncryption()
}
