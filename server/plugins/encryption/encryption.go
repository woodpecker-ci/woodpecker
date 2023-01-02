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
)

const (
	rawKeyConfigFlag             = "encryption-raw-key"
	tinkKeysetFilepathConfigFlag = "encryption-tink-keyset"
	disableEncryptionConfigFlag  = "encryption-disable-flag"

	ciphertextSampleConfigKey = "encryption-ciphertext-sample"

	keyTypeTink = "tink"
	keyTypeRaw  = "raw"
	keyTypeNone = "none"
)

var (
	encryptionNotEnabledError = errors.New("encryption is not enabled")
	encryptionKeyInvalidError = errors.New("encryption key is invalid")
	encryptionKeyRotatedError = errors.New("encryption key is being rotated")
)

type builder struct {
	store   store.Store
	ctx     *cli.Context
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
		_, err := noEncryptionBuilder{}.WithClients(b.clients).Build()
		if err != nil {
			log.Fatal().Err(err).Msg("failed initializing server in unencrypted mode")
		}
		return
	}
	svc := b.getService(keyType)
	if disableFlag {
		err := svc.Disable()
		if err != nil {
			log.Fatal().Err(err).Msg("failed disabling encryption")
		}
	}

}
