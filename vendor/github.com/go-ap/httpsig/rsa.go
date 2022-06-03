// Copyright (C) 2017 Space Monkey, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpsig

import (
	"crypto"
	"crypto/rsa"
)

// RSASHA1 implements RSA PKCS1v15 signatures over a SHA1 digest
var RSASHA1 Algorithm = rsaSha1{}

type rsaSha1 struct{}

func (rsaSha1) Name() string {
	return "rsa-sha1"
}

func (a rsaSha1) Sign(key interface{}, data []byte) ([]byte, error) {
	k := toRSAPrivateKey(key)
	if k == nil {
		return nil, unsupportedAlgorithm(a)
	}
	return RSASign(k, crypto.SHA1, data)
}

func (a rsaSha1) Verify(key interface{}, data, sig []byte) error {
	k := toRSAPublicKey(key)
	if k == nil {
		return unsupportedAlgorithm(a)
	}
	return RSAVerify(k, crypto.SHA1, data, sig)
}

// RSASHA256 implements RSA PKCS1v15 signatures over a SHA256 digest
var RSASHA256 Algorithm = rsaSha256{}

type rsaSha256 struct{}

func (rsaSha256) Name() string {
	return "rsa-sha256"
}

func (a rsaSha256) Sign(key interface{}, data []byte) ([]byte, error) {
	k := toRSAPrivateKey(key)
	if k == nil {
		return nil, unsupportedAlgorithm(a)
	}
	return RSASign(k, crypto.SHA256, data)
}

func (a rsaSha256) Verify(key interface{}, data, sig []byte) error {
	k := toRSAPublicKey(key)
	if k == nil {
		return unsupportedAlgorithm(a)
	}
	return RSAVerify(k, crypto.SHA256, data, sig)
}

// RSASign signs a digest of the data hashed using the provided hash
func RSASign(key *rsa.PrivateKey, hash crypto.Hash, data []byte) (
	signature []byte, err error) {

	h := hash.New()
	if _, err := h.Write(data); err != nil {
		return nil, err
	}
	return rsa.SignPKCS1v15(Rand, key, hash, h.Sum(nil))
}

// RSAVerify verifies a signed digest of the data hashed using the provided hash
func RSAVerify(key *rsa.PublicKey, hash crypto.Hash, data, sig []byte) (
	err error) {

	h := hash.New()
	if _, err := h.Write(data); err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(key, hash, h.Sum(nil), sig)
}
