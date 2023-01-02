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
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

type tinkConfiguration struct {
	keysetFilePath string
	store          store.Store
	clients        []model.EncryptionClient
}

func newTink(ctx *cli.Context, s store.Store) model.EncryptionServiceBuilder {
	filepath := ctx.String(tinkKeysetFilepathConfigFlag)
	return &tinkConfiguration{filepath, s, nil}
}

func (c tinkConfiguration) WithClients(clients []model.EncryptionClient) model.EncryptionServiceBuilder {
	c.clients = clients
	return c
}

func (c tinkConfiguration) Build() (model.EncryptionService, error) {
	svc := &tinkEncryptionService{
		keysetFilePath:    c.keysetFilePath,
		primaryKeyID:      "",
		encryption:        nil,
		store:             c.store,
		keysetFileWatcher: nil,
		clients:           c.clients,
	}
	err := svc.initClients()
	if err != nil {
		return nil, fmt.Errorf("failed initializing encryption clients: %w", err)
	}

	err = svc.loadKeyset()
	if err != nil {
		return nil, fmt.Errorf("failed loading encryption keyset: %w", err)
	}

	err = svc.validateKeyset()
	if err == errEncryptionNotEnabled {
		err = svc.enable()
	} else if err == errEncryptionKeyRotated {
		err = svc.rotate()
	}

	if err != nil {
		return nil, fmt.Errorf("failed validating encryption keyset: %w", err)
	}

	err = svc.initFileWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed initializing keyset file watcher: %w", err)
	}
	return svc, nil
}
