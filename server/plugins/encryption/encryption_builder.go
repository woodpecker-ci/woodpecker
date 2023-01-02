package encryption

import (
	"errors"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store/types"
)

func (b builder) getService(keyType string) (model.EncryptionService, error) {
	if keyType == keyTypeNone {
		return nil, errors.New("encryption enabled but no keys provided")
	}

	builder, err := b.serviceBuilder(keyType)
	if err != nil {
		return nil, err
	}

	svc, err := builder.WithClients(b.clients).Build()
	if err != nil {
		return nil, err
	}
	return svc, nil
}

func (b builder) isEnabled() (bool, error) {
	_, err := b.store.ServerConfigGet(ciphertextSampleConfigKey)
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		return false, errors.New("failed to load encryption configuration")
	}
	return err == nil, nil
}

func (b builder) detectKeyType() (string, error) {
	rawKeyPresent := b.ctx.IsSet(rawKeyConfigFlag)
	tinkKeysetPresent := b.ctx.IsSet(tinkKeysetFilepathConfigFlag)
	if rawKeyPresent && tinkKeysetPresent {
		return "", errors.New("can not use raw encryption key and tink keyset at the same time")
	} else if rawKeyPresent {
		return keyTypeRaw, nil
	} else if tinkKeysetPresent {
		return keyTypeTink, nil
	}
	return keyTypeNone, nil
}

func (b builder) serviceBuilder(keyType string) (model.EncryptionServiceBuilder, error) {
	if keyType == keyTypeTink {
		return newTink(b.ctx, b.store), nil
	} else if keyType == keyTypeRaw {
		return newAES(b.ctx, b.store), nil
	} else if keyType == keyTypeNone {
		return &noEncryptionBuilder{}, nil
	} else {
		return nil, errors.New(fmt.Sprintf("unsupported encryption key type: %s", keyType))
	}
}
