package httpsignatures

import (
	"crypto/sha1"
	"crypto/sha256"
	"errors"
	"hash"
)

var (
	AlgorithmHmacSha256 = &Algorithm{"hmac-sha256", sha256.New}
	AlgorithmHmacSha1   = &Algorithm{"hmac-sha1", sha1.New}

	ErrorUnknownAlgorithm = errors.New("Unknown Algorithm")
)

type Algorithm struct {
	name string
	hash func() hash.Hash
}

func algorithmFromString(name string) (*Algorithm, error) {
	switch name {
	case AlgorithmHmacSha1.name:
		return AlgorithmHmacSha1, nil
	case AlgorithmHmacSha256.name:
		return AlgorithmHmacSha256, nil
	}

	return nil, ErrorUnknownAlgorithm
}
