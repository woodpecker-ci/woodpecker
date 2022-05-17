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
	pubKeyID := "woodpecker-ci-plugins"

	pubEd25519Key, privEd25519Key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	body := []byte("{\"foo\":\"bar\"}")

	req, err := http.NewRequest("GET", "http://example.com", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}

	req.Header.Set("Content-Type", "application/json")

	err = utils.SignHTTPRequest(privEd25519Key, req, body)
	if err != nil {
		t.Fatal(err)
	}

	verifier, err := httpsig.NewVerifier(req)
	if err != nil {
		t.Fatal(err)
	}

	if verifier.KeyId() != pubKeyID {
		t.Fatalf("expected pubKeyId to be %s, got %s", pubKeyID, verifier.KeyId())
	}

	err = verifier.Verify(pubEd25519Key, httpsig.ED25519)
	if err != nil {
		t.Fatal(err)
	}
}
