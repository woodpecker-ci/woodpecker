package encrypted_secrets

import (
	"errors"
	"github.com/google/tink/go/daead"
	"github.com/google/tink/go/insecurecleartextkeyset"
	"github.com/google/tink/go/keyset"
	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server/store/types"
	"os"
	"strconv"
)

const keyIdAAD = "Primary key id"

// Init and hot reload encryption primitive
func (svc *Encryption) initEncryption() {
	log.Warn().Msgf("Loading secrets encryption keyset from file: %s", svc.keysetFilePath)
	file, err := os.Open(svc.keysetFilePath)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error opening secret encryption keyset file")
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Err(err).Msgf("Could not close keyset file: %s", svc.keysetFilePath)
		}
	}(file)

	jsonKeyset := keyset.NewJSONReader(file)
	keysetHandle, err := insecurecleartextkeyset.Read(jsonKeyset)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error reading secret encryption keyset")
	}
	svc.primaryKeyId = strconv.FormatUint(uint64(keysetHandle.KeysetInfo().PrimaryKeyId), 10)

	encryptionInstance, err := daead.New(keysetHandle)
	if err != nil {
		log.Fatal().Err(err).Msgf("Error initializing secret encryption")
	}
	svc.encryption = encryptionInstance

	svc.validateKeyset()
}

// DB ciphertext sample
// store encrypted primaryKeyId in DB to check if used keyset is the same as used to encrypt secrets data
// and to detect keyset rotations
func (svc *Encryption) validateKeyset() {
	ciphertextSample, err := svc.store.ServerConfigGet("secrets-encryption-key-id")
	if errors.Is(err, types.RecordNotExist) {
		svc.updateCiphertextSample()
		log.Warn().Msg("Secrets encryption enabled on server")
		svc.encryptDatabase()
		return
	} else if err != nil {
		log.Fatal().Err(err).Msgf("Invalid secrets encryption key")
	}

	plaintext := svc.decrypt(ciphertextSample, keyIdAAD)
	if err != nil {
		log.Fatal().Err(err).Msgf("Secrets encryption error")
	} else if plaintext != svc.primaryKeyId {
		svc.updateCiphertextSample()
		log.Info().Msg("Registered rotated secrets encryption key")
		svc.reEncryptDatabase()
	}
}

func (svc *Encryption) updateCiphertextSample() {
	ct := svc.encrypt(svc.primaryKeyId, keyIdAAD)

	err := svc.store.ServerConfigSet("secrets-encryption-key-id", ct)
	if err != nil {
		log.Fatal().Err(err).Msgf("Storage error")
	}
}
