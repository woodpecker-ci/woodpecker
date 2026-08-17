// Copyright 2026 Woodpecker Authors
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

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatasourceDefaultValue(t *testing.T) {
	t.Run("outside container", func(t *testing.T) {
		unsetEnvForTest(t, "WOODPECKER_IN_CONTAINER")
		assert.Equal(t, "woodpecker.sqlite", datasourceDefaultValue())
	})

	t.Run("inside container", func(t *testing.T) {
		t.Setenv("WOODPECKER_IN_CONTAINER", "true")
		assert.Equal(t, "/var/lib/woodpecker/woodpecker.sqlite", datasourceDefaultValue())
	})

	t.Run("inside container with empty value still counts as set", func(t *testing.T) {
		// LookupEnv reports an empty-but-present var as found
		t.Setenv("WOODPECKER_IN_CONTAINER", "")
		assert.Equal(t, "/var/lib/woodpecker/woodpecker.sqlite", datasourceDefaultValue())
	})
}

func TestGetFirstNonEmptyEnvVar(t *testing.T) {
	t.Run("returns first set non-empty var", func(t *testing.T) {
		t.Setenv("WP_TEST_A", "")
		t.Setenv("WP_TEST_B", "value-b")
		t.Setenv("WP_TEST_C", "value-c")
		assert.Equal(t, "value-b", getFirstNonEmptyEnvVar("WP_TEST_A", "WP_TEST_B", "WP_TEST_C"))
	})

	t.Run("returns empty when all unset or empty", func(t *testing.T) {
		t.Setenv("WP_TEST_A", "")
		t.Setenv("WP_TEST_B", "")
		assert.Empty(t, getFirstNonEmptyEnvVar("WP_TEST_A", "WP_TEST_B"))
	})

	t.Run("no args", func(t *testing.T) {
		assert.Empty(t, getFirstNonEmptyEnvVar())
	})
}

// unsetEnvForTest removes an env var and returns a function that restores its
// previous value. Needed because t.Setenv cannot represent an *absent* var.
func unsetEnvForTest(t *testing.T, key string) {
	t.Helper()
	prev, had := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, prev) //nolint:usetesting
		} else {
			_ = os.Unsetenv(key)
		}
	})
}
