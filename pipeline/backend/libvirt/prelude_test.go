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

//go:build cgo && libvirt

package libvirt

import (
	"errors"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCmdError(t *testing.T) {
	t.Run("exit error", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "echo broken >&2; exit 3")
		_, err := cmd.Output()
		require.Error(t, err)

		err = cmdError(cmd, err)
		assert.ErrorContains(t, err, "and stderr: broken")
		_, ok := errors.AsType[*exec.ExitError](err)
		assert.True(t, ok)
	})

	t.Run("command not found", func(t *testing.T) {
		cmd := exec.Command("woodpecker-command-that-does-not-exist")
		_, err := cmd.Output()
		require.Error(t, err)

		err = cmdError(cmd, err)
		assert.ErrorContains(t, err, "command woodpecker-command-that-does-not-exist failed with")
		assert.ErrorIs(t, err, exec.ErrNotFound)
	})
}
