package encryption

import (
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store/types"
)

func (b builder) getService(keyType string) model.EncryptionService {
	if keyType == keyTypeNone {
		log.Fatal().Msg("Encryption enabled but no keys provided")
	}
	svc, err := b.serviceBuilder(keyType).WithClients(b.clients).Build()
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing encryption")
	}
	return svc
}

func (b builder) isEnabled() bool {
	_, err := b.store.ServerConfigGet(ciphertextSampleConfigKey)
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		log.Fatal().Err(err).Msg("failed to load encryption configuration")
	}
	return err == nil
}

func (b builder) detectKeyType() string {
	rawKeyPresent := b.ctx.IsSet(rawKeyConfigFlag)
	tinkKeysetPresent := b.ctx.IsSet(tinkKeysetFilepathConfigFlag)
	if rawKeyPresent && tinkKeysetPresent {
		log.Fatal().Msg("Can not use raw encryption key and tink keyset at the same time")
	} else if rawKeyPresent {
		return keyTypeRaw
	} else if tinkKeysetPresent {
		return keyTypeTink
	}
	return keyTypeNone
}

func (b builder) serviceBuilder(keyType string) model.EncryptionServiceBuilder {
	if keyType == keyTypeTink {
		return newTink(b.ctx, b.store)
	} else if keyType == keyTypeRaw {
		return newAES(b.ctx, b.store)
	} else if keyType == keyTypeNone {
		return &noEncryptionBuilder{}
	} else {
		log.Fatal().Msgf("unsupported encryption key type: %s", keyType)
		return nil
	}
}
