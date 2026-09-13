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

package main

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppHasDecodeBase64Command(t *testing.T) {
	for _, command := range newApp().Commands {
		if command.Name == "decode-base64" {
			assert.True(t, command.Hidden)
			return
		}
	}

	t.Fatal("decode-base64 command not found")
}

func TestNewAppRunsDecodeBase64WithoutConfig(t *testing.T) {
	if os.Getenv("WOODPECKER_DECODE_BASE64_HELPER") == "1" {
		err := newApp().Run(context.Background(), []string{"woodpecker-cli", "decode-base64", "KyBlY2hvIGhpCg=="})
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	}

	command := exec.CommandContext(t.Context(), os.Args[0], "-test.run=^TestNewAppRunsDecodeBase64WithoutConfig$")
	command.Env = append(
		os.Environ(),
		"WOODPECKER_DECODE_BASE64_HELPER=1",
		"WOODPECKER_DISABLE_UPDATE_CHECK=true",
		"XDG_CONFIG_HOME="+t.TempDir(),
	)
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
	assert.Equal(t, "+ echo hi\n", string(output))
}
