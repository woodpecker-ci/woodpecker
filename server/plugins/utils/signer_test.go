package utils_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"net/http"
	"testing"

	"github.com/go-fed/httpsig"
	"github.com/woodpecker-ci/woodpecker/server/plugins/utils"
)

func TestSign(t *testing.T) {
	pubKeyId := "pubEd25519Key"

	pubEd25519Key, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	s := utils.NewSigner(privEd25519Key, pubKeyId)

	req, err := http.NewRequest("GET", "http://example.com", bytes.NewBuffer([]byte("{\"foo\":\"bar\"}")))
	if err != nil {
		t.Error(err)
	}

	err = s.Sign(req)
	if err != nil {
		t.Error(err)
	}

	verifier, err := httpsig.NewVerifier(req)
	if err != nil {
		t.Error(err)
	}

	if verifier.KeyId() != pubKeyId {
		t.Errorf("expected pubKeyId to be %s, got %s", pubKeyId, pubKeyId)
	}

	err = verifier.Verify(pubEd25519Key, httpsig.ED25519)
	if err != nil {
		t.Error(err)
	}
}
