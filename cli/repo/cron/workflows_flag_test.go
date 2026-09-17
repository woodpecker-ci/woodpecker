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
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/urfave/cli/v3"
)

// The server keeps a cron's workflow selection when the patch carries no
// "workflows" at all and clears it when the patch carries an empty list, so
// the update command must send nil when the user said nothing about
// workflows and an empty, non-nil slice when they asked to clear.
func TestCronUpdateWorkflowsFlag(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{"omitted leaves the selection alone", nil, nil},
		{"named workflows replace the selection", []string{"--workflow", "deploy", "-w", "test"}, []string{"deploy", "test"}},
		{"--clear-workflows sends an empty list", []string{"--clear-workflows"}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			command := *cronUpdateCmd
			command.Flags = freshFlags(cronUpdateCmd.Flags)
			command.Writer = io.Discard
			command.Action = func(_ context.Context, c *cli.Command) (err error) {
				got, err = cronWorkflowsFromFlags(c)
				return err
			}

			args := append([]string{"update", "--repository", "owner/repo", "--id", "1"}, tt.args...)
			err := command.Run(t.Context(), args)

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			if tt.want != nil {
				assert.NotNil(t, got)
			}
		})
	}
}

func TestCronUpdateRejectsClearAlongsideWorkflow(t *testing.T) {
	command := *cronUpdateCmd
	command.Flags = freshFlags(cronUpdateCmd.Flags)
	command.Writer = io.Discard
	command.Action = func(_ context.Context, c *cli.Command) error {
		_, err := cronWorkflowsFromFlags(c)
		return err
	}

	err := command.Run(t.Context(), []string{"update", "--repository", "owner/repo", "--id", "1", "--clear-workflows", "-w", "deploy"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--clear-workflows")
}

// freshFlags copies every flag so a run cannot leave "has been set" state
// behind on the package-level flag pointers for the next subtest to read.
func freshFlags(flags []cli.Flag) []cli.Flag {
	fresh := make([]cli.Flag, 0, len(flags))
	for _, flag := range flags {
		value := reflect.ValueOf(flag)
		copied := reflect.New(value.Elem().Type())
		copied.Elem().Set(value.Elem())
		fresh = append(fresh, copied.Interface().(cli.Flag))
	}
	return fresh
}
