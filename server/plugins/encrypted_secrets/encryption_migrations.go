package encrypted_secrets

import (
	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server/model"
)

// Encrypt database after encryption was enabled
func (svc *Encryption) encryptDatabase() {
	log.Warn().Msg("Encrypting all secrets in database")
	for _, secret := range svc.fetchAllSecrets() {
		svc.encryptSecret(secret)
		svc.saveSecret(secret)
	}
	log.Warn().Msg("All secrets are encrypted")
}

// Re-encrypt database after key rotations
func (svc *Encryption) reEncryptDatabase() {
	log.Warn().Msg("Re-encrypting all secrets in database")
	for _, secret := range svc.fetchAllSecrets() {
		svc.decryptSecret(secret)
		svc.encryptSecret(secret)
		svc.saveSecret(secret)
	}
	log.Warn().Msg("All secrets are re-encrypted")
}

// Decrypt database
func (svc *Encryption) decryptDatabase() {
	log.Warn().Msg("Decrypting all secrets")
	for _, secret := range svc.fetchAllSecrets() {
		svc.decryptSecret(secret)
		svc.saveSecret(secret)
	}
	log.Warn().Msg("Secrets are decrypted")
}

func (svc *Encryption) fetchAllSecrets() []*model.Secret {
	secrets, err := svc.store.SecretListAll()
	if err != nil {
		log.Fatal().Err(err).Msg("Secrets decryption failed: could not fetch secrets from DB")
	}
	return secrets
}

func (svc *Encryption) saveSecret(secret *model.Secret) {
	err := svc.store.SecretUpdate(secret)
	if err != nil {
		log.Fatal().Err(err).Msg("Storage error: could not update secret in DB")
	}
}
