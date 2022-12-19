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

const tinkKeysetConfigFlag = "secrets-encryption-keyset"

type tinkConfiguration struct {
	keysetFilePath string
	store          store.Store
	clients        []model.EncryptionClient
}

func newTink(ctx *cli.Context, s store.Store) model.EncryptionServiceBuilder {
	filepath := ctx.String(tinkKeysetConfigFlag)
	return &tinkConfiguration{filepath, s, nil}
}

func (c tinkConfiguration) WithClients(clients []model.EncryptionClient) model.EncryptionServiceBuilder {
	c.clients = clients
	return c
}

func (c tinkConfiguration) Build() model.EncryptionService {
	svc := tinkEncryptionService{
		keysetFilePath:    c.keysetFilePath,
		primaryKeyId:      "",
		encryption:        nil,
		store:             c.store,
		keysetFileWatcher: nil,
		clients:           c.clients,
	}
	svc.loadKeyset()
	err := svc.validateKeyset()
	if err == encryptionNotEnabledError {
		svc.enable()
	} else if err == encryptionKeyInvalidError {
		log.Fatal().Err(err)
	} else if err == encryptionKeyRotatedError {
		svc.rotate()
	}
	svc.initFileWatcher()
	return &svc
}
