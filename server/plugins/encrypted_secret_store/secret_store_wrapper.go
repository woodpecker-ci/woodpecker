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

func New(secretStore model.SecretStore) *EncryptedSecretStore {
	wrapper := EncryptedSecretStore{secretStore, nil}
	return &wrapper
}

func (wrapper *EncryptedSecretStore) SetEncryptionService(encryption model.EncryptionService) {
	if wrapper.encryption != nil {
		log.Fatal().Msg("Attempt to init more than once")
	}
	wrapper.encryption = encryption
}

func (wrapper *EncryptedSecretStore) EnableEncryption() {
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

func (wrapper *EncryptedSecretStore) MigrateEncryption(newEncryptionService model.EncryptionService) {
	log.Warn().Msg("Migrating secrets encryption")
	secrets, err := wrapper.store.GlobalSecretList()
	if err != nil {
		log.Fatal().Err(err).Msg("Secrets encryption migration failed: could not fetch secrets from DB")
	}
	wrapper.decryptList(secrets)
	wrapper.encryption = newEncryptionService
	for _, secret := range secrets {
		wrapper.encrypt(secret)
		wrapper._save(secret)
	}
	log.Warn().Msg("Secrets encryption migrated successfully")
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
