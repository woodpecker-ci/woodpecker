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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandomString(t *testing.T) {
	const alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

	for _, length := range []int{1, 7, 8, 32, 52, 100} {
		value := RandomString(length)

		assert.Len(t, value, length)
		assert.NotContains(t, value, "=", "padding would need escaping in URLs")
		for _, char := range value {
			assert.True(t, strings.ContainsRune(alphabet, char), "unexpected character %q", char)
		}
	}
}

func TestRandomStringIsRandom(t *testing.T) {
	const runs = 100

	seen := make(map[string]struct{}, runs)
	for range runs {
		seen[RandomString(52)] = struct{}{}
	}

	assert.Len(t, seen, runs, "every call should return its own value")
}

func TestRandomStringWithoutLength(t *testing.T) {
	assert.Empty(t, RandomString(0))
	assert.Empty(t, RandomString(-1))
}
