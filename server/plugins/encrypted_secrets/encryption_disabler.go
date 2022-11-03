package encrypted_secrets

import (
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

// DecryptAll - Decrypt entire DB and disable secrets encryption
func DecryptAll(ctx *cli.Context, s store.Store) {
	filepath := ctx.String("secrets-encryption-decrypt-all-keyset")

	service := Encryption{s, nil, "", filepath, nil}
	service.initEncryption()
	service.decryptDatabase()
	err := service.store.ServerConfigDelete("secrets-encryption-key-id")
	if err != nil {
		log.Fatal().Err(err).Msg("Disabling secrets encryption failed: could not update server config")
	}
	log.Warn().Msg("Secrets encryption disabled")
}
