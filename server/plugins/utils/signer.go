package utils

import (
	"crypto/ed25519"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-fed/httpsig"
)

type signer struct {
	privateKey  ed25519.PrivateKey
	publicKeyId string
}

func NewSigner(privateKey ed25519.PrivateKey, publicKeyId string) *signer {
	return &signer{
		privateKey:  privateKey,
		publicKeyId: publicKeyId,
	}
}

func (s *signer) Sign(req *http.Request) error {
	prefs := []httpsig.Algorithm{httpsig.ED25519}
	headers := []string{httpsig.RequestTarget, "date"}
	signer, _, err := httpsig.NewSigner(prefs, httpsig.DigestSha256, headers, httpsig.Signature, 0)
	if err != nil {
		return err
	}

	req.Header.Add("date", time.Now().UTC().Format(http.TimeFormat))

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return err
	}

	if err = signer.SignRequest(s.privateKey, s.publicKeyId, req, body); err != nil {
		return err
	}

	return nil
}
