package encryption

import (
	"github.com/google/tink/go/subtle/random"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestShortEncryptDecrypt(t *testing.T) {
	aes := &aesEncryptionService{}
	aes.loadCipher([]byte("eThWmZq4t7w!z%C*F-JaNdRfUjXn2r5u"))
	input := string(random.GetRandomBytes(4))
	cipher := aes.Encrypt(input, "")
	output := aes.Decrypt(cipher, "")
	assert.Equal(t, input, output)
}

func TestLongEncryptDecrypt(t *testing.T) {
	aes := &aesEncryptionService{}
	aes.loadCipher([]byte("eThWmZq4t7w!z%C*F-JaNdRfUjXn2r5u"))
	input := string(random.GetRandomBytes(1024))
	cipher := aes.Encrypt(input, "")
	output := aes.Decrypt(cipher, "")
	assert.Equal(t, input, output)
}
