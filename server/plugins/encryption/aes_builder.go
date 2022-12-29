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

type aesConfiguration struct {
	key     string
	store   store.Store
	clients []model.EncryptionClient
}

func newAES(ctx *cli.Context, s store.Store) model.EncryptionServiceBuilder {
	key := ctx.String(rawKeyConfigFlag)
	return &aesConfiguration{key, s, nil}
}

func (c aesConfiguration) WithClients(clients []model.EncryptionClient) model.EncryptionServiceBuilder {
	c.clients = clients
	return c
}

func (c aesConfiguration) Build() model.EncryptionService {
	svc := &aesEncryptionService{
		cipher:  nil,
		store:   c.store,
		clients: c.clients,
	}
	svc.initClients()
	svc.loadCipher([]byte(c.key))
	err := svc.validateCipher()
	if err == encryptionNotEnabledError {
		svc.enable()
	} else if err == encryptionKeyInvalidError {
		log.Fatal().Err(err).Msg("Error initializing AES encryption")
	}
	return svc
}
