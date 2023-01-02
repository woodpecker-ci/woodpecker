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
	"fmt"

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

	keyIDAssociatedData = "Primary key id"
)

var (
	errEncryptionNotEnabled = errors.New("encryption is not enabled")
	errEncryptionKeyInvalid = errors.New("encryption key is invalid")
	errEncryptionKeyRotated = errors.New("encryption key is being rotated")
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

func (b builder) Build() error {
	enabled, err := b.isEnabled()
	if err != nil {
		return fmt.Errorf("failed fetching server encryption status: %w", err)
	}

	disableFlag := b.ctx.Bool(disableEncryptionConfigFlag)

	keyType, err := b.detectKeyType()
	if err != nil {
		return fmt.Errorf("failed determining encryption key type: %w", err)
	}

	if !enabled && (disableFlag || keyType == keyTypeNone) {
		_, err := noEncryptionBuilder{}.WithClients(b.clients).Build()
		if err != nil {
			return fmt.Errorf("failed initializing server in unencrypted mode: %w", err)
		}
	}
	svc, err := b.getService(keyType)
	if err != nil {
		return fmt.Errorf("failed initializing encryption service: %w", err)
	}

	if disableFlag {
		err := svc.Disable()
		if err != nil {
			return fmt.Errorf("failed disabling encryption: %w", err)
		}
	}
	return nil
}
