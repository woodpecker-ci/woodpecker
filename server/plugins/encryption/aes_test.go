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
	"testing"

	"github.com/google/tink/go/subtle/random"
	"github.com/stretchr/testify/assert"
)

func TestEncryptDecryptShortMessage(t *testing.T) {
	aes := &aesEncryptionService{}
	aes.loadCipher(random.GetRandomBytes(32))
	input := string(random.GetRandomBytes(4))
	cipher := aes.Encrypt(input, "")
	output := aes.Decrypt(cipher, "")
	assert.Equal(t, input, output)
}

func TestEncryptDecryptLongMessage(t *testing.T) {
	aes := &aesEncryptionService{}
	aes.loadCipher(random.GetRandomBytes(32))
	input := string(random.GetRandomBytes(1024))
	cipher := aes.Encrypt(input, "")
	output := aes.Decrypt(cipher, "")
	assert.Equal(t, input, output)
}
