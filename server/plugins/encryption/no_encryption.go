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

func (svc *noEncryption) Encrypt(plaintext string, _ string) string {
	return plaintext
}

func (svc *noEncryption) Decrypt(ciphertext string, _ string) string {
	return ciphertext
}

func (svc *noEncryption) Disable() {
	return
}
