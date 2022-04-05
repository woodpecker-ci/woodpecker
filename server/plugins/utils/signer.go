package utils

import (
	"crypto/ed25519"
	"net/http"
	"time"

	"github.com/go-fed/httpsig"
)

type signer struct {
	privateKey  ed25519.PrivateKey
	publicKeyId string
}

func (s *signer) sign(req *http.Request) error {
	prefs := []httpsig.Algorithm{httpsig.RSA_SHA256}
	headers := []string{httpsig.RequestTarget, "date"}
	signer, _, err := httpsig.NewSigner(prefs, httpsig.DigestSha256, headers, httpsig.Signature, 0)
	if err != nil {
		return err
	}

	req.Header.Add("date", time.Now().UTC().Format(http.TimeFormat))

	if err = signer.SignRequest(s.privateKey, s.publicKeyId, req, jsonBytes); err != nil {
		return err
	}

	return nil
}
