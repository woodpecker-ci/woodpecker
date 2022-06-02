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

// Algorithm provides methods used to sign/verify signatures.
type Algorithm interface {
	Name() string
	Sign(key interface{}, data []byte) (sig []byte, err error)
	Verify(key interface{}, data, sig []byte) error
}

// KeyGetter is an interface used by the verifier to retrieve a key stored
// by key id.
//
// The following types are supported for the specified algorithms:
// []byte            - HMAC signatures
// *rsa.PublicKey    - RSA signatures
// *rsa.PrivateKey   - RSA signatures
//
// Other types will treated as if no key was returned.
type KeyGetter interface {
	GetKey(id string) (interface{}, error)
}

// KeyGetterFunc is a convenience type for implementing a KeyGetter with a
// regular function
type KeyGetterFunc func(id string) (interface{}, error)

// GetKey calls fn(id)
func (fn KeyGetterFunc) GetKey(id string) (interface{}, error) {
	return fn(id)
}
