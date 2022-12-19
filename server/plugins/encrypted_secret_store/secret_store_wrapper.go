package encrypted_secret_store

import (
	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"strconv"
)

type EncryptedSecretStore struct {
	store      model.SecretStore
	encryption model.EncryptionService
}

func New(secretStore model.SecretStore) model.SecretStore {
	wrapper := EncryptedSecretStore{secretStore, nil}
	return &wrapper
}

func (wrapper *EncryptedSecretStore) InitEncryption(encryption model.EncryptionService) {
	if wrapper.encryption != nil {
		log.Fatal().Msg("Attempt to init more than once")
	}
	wrapper.encryption = encryption
}

func (wrapper *EncryptedSecretStore) EncryptStore() {
	log.Warn().Msg("Encrypting all secrets in database")
	secrets, err := wrapper.store.GlobalSecretList()
	if err != nil {
		log.Fatal().Err(err).Msg("Secrets encryption failed: could not fetch secrets from DB")
	}
	for _, secret := range secrets {
		wrapper.encrypt(secret)
		wrapper._save(secret)
	}
	log.Warn().Msg("All secrets are encrypted")
}

func (wrapper *EncryptedSecretStore) ReEncryptStore(newEncryptionService model.EncryptionService) {
	log.Warn().Msg("Re-encrypting all secrets in database")
	secrets, err := wrapper.store.GlobalSecretList()
	if err != nil {
		log.Fatal().Err(err).Msg("Secrets key rotation failed: could not fetch secrets from DB")
	}
	wrapper.decryptList(secrets)
	wrapper.encryption = newEncryptionService
	for _, secret := range secrets {
		wrapper.encrypt(secret)
		wrapper._save(secret)
	}
	log.Warn().Msg("All secrets are re-encrypted")
}

func (wrapper *EncryptedSecretStore) DecryptStore() {
	log.Warn().Msg("Decrypting all secrets")
	secrets, err := wrapper.store.GlobalSecretList()
	if err != nil {
		log.Fatal().Err(err).Msg("Secrets decryption failed: could not fetch secrets from DB")
	}
	for _, secret := range secrets {
		wrapper.decrypt(secret)
		wrapper._save(secret)
	}
	log.Warn().Msg("Secrets are decrypted")
}

func (wrapper *EncryptedSecretStore) encrypt(secret *model.Secret) {
	encryptedValue := wrapper.encryption.Encrypt(secret.Value, strconv.Itoa(int(secret.ID)))
	secret.Value = encryptedValue
}

func (wrapper *EncryptedSecretStore) decrypt(secret *model.Secret) {
	decryptedValue := wrapper.encryption.Decrypt(secret.Value, strconv.Itoa(int(secret.ID)))
	secret.Value = decryptedValue
}

func (wrapper *EncryptedSecretStore) decryptList(secrets []*model.Secret) {
	for _, secret := range secrets {
		wrapper.decrypt(secret)
	}
}

func (wrapper *EncryptedSecretStore) _save(secret *model.Secret) {
	err := wrapper.store.SecretUpdate(secret)
	if err != nil {
		log.Fatal().Err(err).Msg("Storage error: could not update secret in DB")
	}
}
