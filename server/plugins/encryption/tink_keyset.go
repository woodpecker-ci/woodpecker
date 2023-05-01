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
	"os"
	"strconv"

	"github.com/google/tink/go/aead"
	"github.com/google/tink/go/insecurecleartextkeyset"
	"github.com/google/tink/go/keyset"
	"github.com/rs/zerolog/log"

	"github.com/woodpecker-ci/woodpecker/server/store/types"
)

func (svc *tinkEncryptionService) loadKeyset() error {
	log.Warn().Msgf(logTemplateTinkLoadingKeyset, svc.keysetFilePath)
	file, err := os.Open(svc.keysetFilePath)
	if err != nil {
		return fmt.Errorf(errTemplateTinkFailedOpeningKeyset, err)
	}
	defer func(file *os.File) {
		err = file.Close()
		if err != nil {
			log.Err(err).Msgf(logTemplateTinkFailedClosingKeysetFile, svc.keysetFilePath)
		}
	}(file)

	jsonKeyset := keyset.NewJSONReader(file)
	keysetHandle, err := insecurecleartextkeyset.Read(jsonKeyset)
	if err != nil {
		return fmt.Errorf(errTemplateTinkFailedReadingKeyset, err)
	}
	svc.primaryKeyID = strconv.FormatUint(uint64(keysetHandle.KeysetInfo().PrimaryKeyId), 10)

	encryptionInstance, err := aead.New(keysetHandle)
	if err != nil {
		return fmt.Errorf(errTemplateTinkFailedInitializingAEAD, err)
	}
	svc.encryption = encryptionInstance
	return nil
}

func (svc *tinkEncryptionService) validateKeyset() error {
	ciphertextSample, err := svc.store.ServerConfigGet(ciphertextSampleConfigKey)
	if errors.Is(err, types.RecordNotExist) {
		return errEncryptionNotEnabled
	} else if err != nil {
		return fmt.Errorf(errTemplateFailedLoadingServerConfig, err)
	}

	plaintext, err := svc.Decrypt(ciphertextSample, keyIDAssociatedData)
	if plaintext != svc.primaryKeyID {
		return errEncryptionKeyRotated
	} else if err != nil {
		return fmt.Errorf(errTemplateFailedValidatingKey, err)
	}
	return nil
}
