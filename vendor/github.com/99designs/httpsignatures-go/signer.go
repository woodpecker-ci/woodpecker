package httpsignatures

import (
	"net/http"
	"strings"
	"time"
)

// Signer is used to create a signature for a given request.
type Signer struct {
	algorithm *Algorithm
	headers   HeaderList
}

var (
	// DefaultSha1Signer will sign requests with the url and date using the SHA1 algorithm.
	// Users are encouraged to create their own signer with the headers they require.
	DefaultSha1Signer = NewSigner(AlgorithmHmacSha1, RequestTarget, "date")

	// DefaultSha256Signer will sign requests with the url and date using the SHA256 algorithm.
	// Users are encouraged to create their own signer with the headers they require.
	DefaultSha256Signer = NewSigner(AlgorithmHmacSha256, RequestTarget, "date")
)

func NewSigner(algorithm *Algorithm, headers ...string) *Signer {
	hl := HeaderList{}

	for _, header := range headers {
		hl = append(hl, strings.ToLower(header))
	}

	return &Signer{
		algorithm: algorithm,
		headers:   hl,
	}
}

// SignRequest adds a http signature to the Signature: HTTP Header
func (s Signer) SignRequest(id, key string, r *http.Request) error {
	sig, err := s.buildSignature(id, key, r)
	if err != nil {
		return err
	}

	r.Header.Add(headerSignature, sig.String())

	return nil
}

// AuthRequest adds a http signature to the Authorization: HTTP Header
func (s Signer) AuthRequest(id, key string, r *http.Request) error {
	sig, err := s.buildSignature(id, key, r)
	if err != nil {
		return err
	}

	r.Header.Add(headerAuthorization, authScheme+sig.String())

	return nil
}

func (s Signer) buildSignature(id, key string, r *http.Request) (*Signature, error) {
	if r.Header.Get("date") == "" {
		r.Header.Set("date", time.Now().Format(time.RFC1123))
	}

	sig := &Signature{
		KeyID:     id,
		Algorithm: s.algorithm,
		Headers:   s.headers,
	}

	err := sig.sign(key, r)
	if err != nil {
		return nil, err
	}

	return sig, nil
}
