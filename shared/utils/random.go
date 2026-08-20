// Copyright 2026 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"encoding/base32"
	"math"

	"github.com/tink-crypto/tink-go/v2/subtle/random"
)

// base32NoPadding keeps generated strings free of the '=' padding, so they can
// be used in URLs and headers without escaping.
var base32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

const (
	// Every base32 character carries five bits.
	bitsPerBase32Char = 5
	bitsPerByte       = 8
)

// RandomString returns a cryptographically secure random string of exactly
// length characters, drawn from the base32 alphabet. It is the single place
// tokens, secrets and hashes are generated, so their strength does not depend
// on every call site getting the encoding right.
func RandomString(length int) string {
	if length <= 0 {
		return ""
	}

	byteLen := (length*bitsPerBase32Char + bitsPerByte - 1) / bitsPerByte
	if byteLen > math.MaxUint32 {
		byteLen = math.MaxUint32
	}

	return base32NoPadding.EncodeToString(random.GetRandomBytes(uint32(byteLen)))[:length]
}
