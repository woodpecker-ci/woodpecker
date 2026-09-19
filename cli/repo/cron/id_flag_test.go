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

package cron

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// --id was a StringFlag on these three commands while every Action read it
// back with c.Int64("id"), which urfave/cli v3 only resolves for a flag
// actually registered as an Int64Flag: for a StringFlag it silently returns
// 0. Every `repo cron show|update|rm --id <n>` therefore always acted on
// cron 0 and 404ed, regardless of the id given on the command line.
func TestCronCommandsParseIDAsInt64(t *testing.T) {
	tests := []struct {
		name string
		cmd  *cli.Command
	}{
		{"show", cronShowCmd},
		{"update", cronUpdateCmd},
		{"rm", cronDeleteCmd},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotID int64
			command := *tt.cmd
			command.Writer = io.Discard
			command.Action = func(_ context.Context, c *cli.Command) error {
				gotID = c.Int64("id")
				return nil
			}

			err := command.Run(t.Context(), []string{tt.name, "--repository", "owner/repo", "--id", "42"})

			require.NoError(t, err)
			assert.Equal(t, int64(42), gotID)
		})
	}
}
