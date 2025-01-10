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

package encryption

import (
	"fmt"

	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/server/services/encryption/types"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
)

type builder struct {
	store   store.Store
	c       *cli.Command
	clients []types.EncryptionClient
}

func Encryption(c *cli.Command, s store.Store) types.EncryptionBuilder {
	return &builder{store: s, c: c}
}

func (b builder) WithClient(client types.EncryptionClient) types.EncryptionBuilder {
	b.clients = append(b.clients, client)
	return b
}

func (b builder) Build() error {
	enabled, err := b.isEnabled()
	if err != nil {
		return err
	}

	disableFlag := b.c.Bool(disableEncryptionConfigFlag)

	keyType, err := b.detectKeyType()
	if err != nil {
		return err
	}

	if !enabled && (disableFlag || keyType == keyTypeNone) {
		_, err := noEncryptionBuilder{}.WithClients(b.clients).Build()
		if err != nil {
			return fmt.Errorf(errTemplateFailedInitializingUnencrypted, err)
		}
	}
	svc, err := b.getService(keyType)
	if err != nil {
		return fmt.Errorf(errTemplateFailedInitializing, err)
	}

	if disableFlag {
		err := svc.Disable()
		if err != nil {
			return err
		}
	}
	return nil
}
