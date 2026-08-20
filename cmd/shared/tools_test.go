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

package shared

import (
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

func TestBase64Decoder(t *testing.T) {
	reader, writer, err := os.Pipe()
	require.NoError(t, err)

	stdout := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() {
		os.Stdout = stdout
		_ = reader.Close()
		_ = writer.Close()
	})

	command := &cli.Command{
		Name:   "decode-base64",
		Action: Base64Decoder,
	}
	require.NoError(t, command.Run(context.Background(), []string{"decode-base64", "KyBlY2hvIGhpCg=="}))
	require.NoError(t, writer.Close())

	output, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, "+ echo hi\n", string(output))
}

func TestBase64DecoderErrors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		message string
	}{
		{
			name:    "missing argument",
			message: "expected exactly one base64 argument",
		},
		{
			name:    "too many arguments",
			args:    []string{"aGVsbG8=", "d29ybGQ="},
			message: "expected exactly one base64 argument",
		},
		{
			name:    "invalid base64",
			args:    []string{"%%%"},
			message: "decode base64: illegal base64 data at input byte 0",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			command := &cli.Command{
				Name:   "decode-base64",
				Action: Base64Decoder,
			}
			err := command.Run(context.Background(), append([]string{"decode-base64"}, test.args...))
			require.EqualError(t, err, test.message)
		})
	}
}
