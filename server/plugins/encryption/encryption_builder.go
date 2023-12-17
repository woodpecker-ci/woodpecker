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
	"errors"
	"fmt"

	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
)

func (b builder) getService(keyType string) (model.EncryptionService, error) {
	if keyType == keyTypeNone {
		return nil, errors.New(errMessageNoKeysProvided)
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
		return false, fmt.Errorf(errTemplateFailedLoadingServerConfig, err)
	}
	return err == nil, nil
}

func (b builder) detectKeyType() (string, error) {
	rawKeyPresent := b.ctx.IsSet(rawKeyConfigFlag)
	tinkKeysetPresent := b.ctx.IsSet(tinkKeysetFilepathConfigFlag)
	switch {
	case rawKeyPresent && tinkKeysetPresent:
		return "", errors.New(errMessageCantUseBothServices)
	case rawKeyPresent:
		return keyTypeRaw, nil
	case tinkKeysetPresent:
		return keyTypeTink, nil
	}
	return keyTypeNone, nil
}

func (b builder) serviceBuilder(keyType string) (model.EncryptionServiceBuilder, error) {
	switch {
	case keyType == keyTypeTink:
		return newTink(b.ctx, b.store), nil
	case keyType == keyTypeRaw:
		return newAES(b.ctx, b.store), nil
	case keyType == keyTypeNone:
		return &noEncryptionBuilder{}, nil
	}
	return nil, fmt.Errorf(errMessageTemplateUnsupportedKeyType, keyType)
}
