package encryption

import (
	"github.com/google/tink/go/subtle/random"
	"github.com/stretchr/testify/assert"
	"testing"
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
