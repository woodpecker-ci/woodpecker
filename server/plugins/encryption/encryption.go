// Copyright 2022 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package encryption

import (
	"errors"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/store/types"
)

const (
	rawKeyConfigFlag             = "encryption-raw-key"
	tinkKeysetFilepathConfigFlag = "encryption-tink-keyset"
	disableEncryptionConfigFlag  = "encryption-disable-flag"

	ciphertextSampleConfigKey = "encryption-ciphertext-sample"

	keyTypeTink = "tink"
	keyTypeRaw  = "raw"
	keyTypeNone = "none"

	encryptionNotEnabledError = encryptionError("encryption is not enabled")
	encryptionKeyInvalidError = encryptionError("encryption key is invalid")
	encryptionKeyRotatedError = encryptionError("encryption key is being rotated")
)

type builder struct {
	store   store.Store
	ctx     *cli.Context
	service model.EncryptionServiceBuilder
	clients []model.EncryptionClient
}

func Encryption(ctx *cli.Context, s store.Store) model.EncryptionBuilder {
	return &builder{store: s, ctx: ctx}
}

func (b builder) WithClient(client model.EncryptionClient) model.EncryptionBuilder {
	b.clients = append(b.clients, client)
	return b
}

func (b builder) Build() {
	enabled := b.isEnabled()
	disableFlag := b.ctx.Bool(disableEncryptionConfigFlag)
	keyType := b.detectKeyType()

	if !enabled && (disableFlag || keyType == keyTypeNone) {
		noEncryptionBuilder{}.WithClients(b.clients).Build()
		return
	}
	svc := b.getService(keyType)
	if disableFlag {
		svc.Disable()
	}
}

func (b builder) getService(keyType string) model.EncryptionService {
	if keyType == keyTypeNone {
		log.Fatal().Msg("Encryption enabled but no keys provided")
	}
	return b.serviceBuilder(keyType).WithClients(b.clients).Build()
}

func (b builder) isEnabled() bool {
	_, err := b.store.ServerConfigGet(ciphertextSampleConfigKey)
	if err != nil && !errors.Is(err, types.RecordNotExist) {
		log.Fatal().Msgf("Failed to read server configuration: %s", err)
	}
	return !errors.Is(err, types.RecordNotExist)
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

type encryptionError string

func (e encryptionError) Error() string { return string(e) }
