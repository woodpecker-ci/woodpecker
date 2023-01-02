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
	"crypto/cipher"
	"encoding/base64"
	"fmt"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/store"
)

type aesEncryptionService struct {
	cipher  cipher.Block
	keyID   string
	store   store.Store
	clients []model.EncryptionClient
}

func (svc *aesEncryptionService) Encrypt(plaintext, _ string) (string, error) {
	msg := []byte(plaintext)
	chainSize := svc.blockSize()
	infoBlock := svc.newSizeInfoChunk(len(msg))
	msg = svc.alignDataByChainSize(msg)
	encrypted := make([]byte, len(infoBlock)+len(msg))
	err := svc.encode(encrypted[0:len(infoBlock)], infoBlock)
	if err != nil {
		return "", fmt.Errorf(errTemplateEncryptionFailed, err)
	}

	for n := 0; n < len(msg)/chainSize; n++ {
		dst, src := encrypted[(n+1)*chainSize:(n+2)*chainSize], msg[n*chainSize:(n+1)*chainSize]
		err = svc.encode(dst, src)
		if err != nil {
			return "", fmt.Errorf(errTemplateEncryptionFailed, err)
		}
	}
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

func (svc *aesEncryptionService) Decrypt(ciphertext, _ string) (string, error) {
	ct, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf(errTemplateBase64DecryptionFailed, err)
	}

	chainSize := svc.blockSize()
	decrypted := make([]byte, len(ct))
	for n := 0; n < len(ct)/chainSize; n++ {
		dst, src := decrypted[n*chainSize:(n+1)*chainSize], ct[n*chainSize:(n+1)*chainSize]
		err = svc.decode(dst, src)
		if err != nil {
			return "", fmt.Errorf(errTemplateDecryptionFailed, err)
		}
	}

	dataLen, err := svc.getDataSize(decrypted)
	if err != nil {
		return "", fmt.Errorf(errTemplateDecryptionFailed, err)
	}
	return string(decrypted[chainSize : chainSize+dataLen]), nil
}

func (svc *aesEncryptionService) Disable() error {
	return svc.disable()
}
