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
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

type Error string

func (e Error) Error() string { return string(e) }

const encryptionNotEnabledError = Error("encryption is not enabled")
const encryptionKeyInvalidError = Error("encryption key is invalid")
const encryptionKeyRotatedError = Error("encryption key is being rotated")

type builder struct {
	store   store.Store
	ctx     *cli.Context
	service model.EncryptionServiceBuilder
	clients []model.EncryptionClient
}

func Encryption(ctx *cli.Context, s store.Store) model.EncryptionBuilder {
	return &builder{store: s, ctx: ctx}
}

func (b builder) OfType(encryptionType string) model.EncryptionBuilder {
	if b.service != nil {
		log.Fatal().Msg("invalid encryption configuration flow: attempt to set encryption type more than once")
	}
	if encryptionType == model.TinkEncryptionType {
		b.service = newTink(b.ctx, b.store)
		return b
	}
	log.Fatal().Msgf("invalid encryption configuration flow: unknown encryption type %s", encryptionType)
	return nil
}

func (b builder) WithClient(client model.EncryptionClient) model.EncryptionBuilder {
	b.clients = append(b.clients, client)
	return b
}

func (b builder) Init() model.EncryptionService {
	return b.service.WithClients(b.clients).Build()
}
