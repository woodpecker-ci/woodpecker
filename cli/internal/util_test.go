// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseKeyPair(t *testing.T) {
	s := []string{"FOO=bar", "BAR=", "BAZ=qux=quux", "INVALID"}
	p := ParseKeyPair(s)
	assert.Equal(t, "bar", p["FOO"])
	assert.Equal(t, "qux=quux", p["BAZ"])
	val, exists := p["BAR"]
	assert.Empty(t, val)
	assert.True(t, exists, "missing a key with no value, keys with empty values are also valid")
	_, exists = p["INVALID"]
	assert.False(t, exists, "keys without an equal sign suffix are invalid")
}
