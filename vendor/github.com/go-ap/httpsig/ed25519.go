package httpsig

import (
	ed "crypto/ed25519"
	"fmt"
)

// Ed25519 implements Ed25519 Algorithm
var Ed25519 Algorithm = ed25519{}

type ed25519 struct{}

func (ed25519) Name() string {
	return "ed25519"
}

func (a ed25519) Sign(key interface{}, data []byte) ([]byte, error) {
	k := toEd25519PrivateKey(key)
	if k == nil {
		return nil, unsupportedAlgorithm(a)
	}
	return Ed25519Sign(k, data)
}

func (a ed25519) Verify(key interface{}, data, sig []byte) error {
	k := toHMACKey(key)
	if k == nil {
		return unsupportedAlgorithm(a)
	}
	return Ed25519Verify(k, data, sig)
}

// Ed25519Verify reports whether sig is a valid signature of message by publicKey.
func Ed25519Verify(key interface{}, message, sig []byte) error {
	k, ok := key.(ed.PublicKey)
	if !ok {
		return fmt.Errorf("key must be an instance of crypto/ed25519.PublicKey")
	}
	if len(k) != ed.PublicKeySize {
		return fmt.Errorf("public key has the wrong size")
	}
	if !ed.Verify(k, message, sig) {
		return fmt.Errorf("signature verification failed")
	}
	return nil
}

// Ed25519Sign signs the message with privateKey and returns a signature.
func Ed25519Sign(key interface{}, message []byte) ([]byte, error) {
	k, ok := key.(ed.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("key must be an instance of crypto/ed25519.PrivateKey")
	}
	if len(k) != ed.PrivateKeySize {
		return nil, fmt.Errorf("private key has the wrong size")
	}
	return ed.Sign(k, message), nil
}
