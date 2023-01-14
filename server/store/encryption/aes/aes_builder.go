// Copyright 2023 Woodpecker Authors
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

package aes

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

type aesConfiguration struct {
	password string
	store    store.Store
	clients  []model.EncryptionClient
}

func newAES(ctx *cli.Context, s store.Store) model.EncryptionServiceBuilder {
	key := ctx.String(rawKeyConfigFlag)
	return &aesConfiguration{key, s, nil}
}

func (c aesConfiguration) WithClients(clients []model.EncryptionClient) model.EncryptionServiceBuilder {
	c.clients = clients
	return c
}

func (c aesConfiguration) Build() (model.EncryptionService, error) {
	svc := &aesEncryptionService{
		cipher:  nil,
		store:   c.store,
		clients: c.clients,
	}
	err := svc.initClients()
	if err != nil {
		return nil, fmt.Errorf(errTemplateFailedInitializingClients, err)
	}

	err = svc.loadCipher(c.password)
	if err != nil {
		return nil, fmt.Errorf(errTemplateAesFailedLoadingCipher, err)
	}

	err = svc.validateKey()
	if err == errEncryptionNotEnabled {
		err = svc.enable()
	}
	if err != nil {
		return nil, fmt.Errorf(errTemplateFailedValidatingKey, err)
	}
	return svc, nil
}
