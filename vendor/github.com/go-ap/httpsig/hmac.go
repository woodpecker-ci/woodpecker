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
	"crypto/hmac"
	"errors"
)

// HMACSHA256 implements keyed HMAC over SHA256 digests
var HMACSHA256 Algorithm = hmacSha256{}

type hmacSha256 struct{}

func (hmacSha256) Name() string {
	return "hmac-sha256"
}

func (a hmacSha256) Sign(key interface{}, data []byte) ([]byte, error) {
	k := toHMACKey(key)
	if k == nil {
		return nil, unsupportedAlgorithm(a)
	}
	return HMACSign(k, crypto.SHA256, data)
}

func (a hmacSha256) Verify(key interface{}, data, sig []byte) error {
	k := toHMACKey(key)
	if k == nil {
		return unsupportedAlgorithm(a)
	}
	return HMACVerify(k, crypto.SHA256, data, sig)
}

// HMACSign signs a digest of the data hashed using the provided hash and key.
func HMACSign(key []byte, hash crypto.Hash, data []byte) ([]byte, error) {
	h := hmac.New(hash.New, key)
	if _, err := h.Write(data); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}

// HMACVerify verifies a signed digest of the data hashed using the provided
// hash and key.
func HMACVerify(key []byte, hash crypto.Hash, data, sig []byte) error {
	actualSig, err := HMACSign(key, hash, data)
	if err != nil {
		return err
	}
	if !hmac.Equal(actualSig, sig) {
		return errors.New("hmac signature mismatch")
	}
	return nil
}
