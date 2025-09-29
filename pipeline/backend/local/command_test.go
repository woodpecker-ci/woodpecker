// Copyright 2025 Woodpecker Authors
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

package local

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenCmdByShell(t *testing.T) {
	tmpDir := t.TempDir()
	e := local{tempDir: tmpDir}

	t.Run("error cases", func(t *testing.T) {
		args, err := e.genCmdByShell("", []string{"echo hi"})
		assert.Nil(t, args)
		assert.ErrorIs(t, err, ErrNoShellSet)

		args, err = e.genCmdByShell("sh", []string{})
		assert.Nil(t, args)
		assert.ErrorIs(t, err, ErrNoCmdSet)
	})

	t.Run("windows shells", func(t *testing.T) {
		t.Run("cmd", func(t *testing.T) {
			args, err := e.genCmdByShell("cmd.exe", []string{"echo hi", "call build.bat"})
			require.NoError(t, err)
			require.Len(t, args, 2)
			assert.Equal(t, "/c", args[0])
			assert.True(t, strings.HasSuffix(args[1], ".cmd"))

			// Verify the temp file was created and contains expected content
			content, err := os.ReadFile(args[1])
			require.NoError(t, err)
			assert.EqualValues(t, `@SET PROMPT=$
@echo + 'echo hi'
@echo hi
@IF NOT %ERRORLEVEL% == 0 exit %ERRORLEVEL%
@echo + 'call build.bat'
@call build.bat
@IF NOT %ERRORLEVEL% == 0 exit %ERRORLEVEL%
`, string(content))
		})

		t.Run("powershell", func(t *testing.T) {
			args, err := e.genCmdByShell("powershell", []string{"Write-Host 'test'", "echo test"})
			require.NoError(t, err)
			require.Len(t, args, 4)
			assert.EqualValues(t, []string{"-noprofile", "-noninteractive", "-c"}, []string{args[0], args[1], args[2]})
			assert.EqualValues(t, `$ErrorActionPreference = "Stop"; echo '+ Write-Host '"'"'test'"'"''
Write-Host 'test'
echo '+ echo test'
echo test`, args[3])

			args, err = e.genCmdByShell("pwsh", []string{"Get-Process"})
			require.NoError(t, err)
			assert.Len(t, args, 4)
			assert.Equal(t, "-noprofile", args[0])
		})
	})

	t.Run("unix shells", func(t *testing.T) {
		args, err := e.genCmdByShell("sh", []string{"echo hello", "pwd"})
		require.NoError(t, err)
		assert.Len(t, args, 3)
		assert.Equal(t, "-e", args[0])
		assert.Equal(t, "-c", args[1])
		assert.Contains(t, args[2], "echo hello")
		assert.Contains(t, args[2], "pwd")

		args, err = e.genCmdByShell("bash", []string{"ls -la"})
		require.NoError(t, err)
		assert.Len(t, args, 3)
		assert.Equal(t, "-e", args[0])
		assert.Equal(t, "-c", args[1])

		args, err = e.genCmdByShell("zsh", []string{"echo test"})
		require.NoError(t, err)
		assert.Len(t, args, 3)
		assert.Equal(t, "-e", args[0])
	})

	t.Run("fish shell", func(t *testing.T) {
		args, err := e.genCmdByShell("fish", []string{"echo test", "ls"})
		require.NoError(t, err)
		assert.Len(t, args, 2)
		assert.Equal(t, "-c", args[0])
		assert.Contains(t, args[1], "echo test")
		assert.Contains(t, args[1], "|| exit $status")
	})

	t.Run("nu shell", func(t *testing.T) {
		args, err := e.genCmdByShell("nu", []string{"echo test"})
		require.NoError(t, err)
		assert.Len(t, args, 2)
		assert.Equal(t, "--commands", args[0])
		assert.Contains(t, args[1], "echo test")
	})

	t.Run("unknown posix shell", func(t *testing.T) {
		// This should trigger probeShellIsPosix which will likely fail for non-existent shell
		args, err := e.genCmdByShell("nonexistentshell", []string{"echo test"})
		if err != nil {
			assert.ErrorIs(t, err, ErrNoPosixShell)
		} else {
			// If somehow it passes, verify it generates posix-style args
			assert.Len(t, args, 3)
			assert.Equal(t, "-e", args[0])
		}
	})

	t.Run("command escaping", func(t *testing.T) {
		args, err := e.genCmdByShell("cmd", []string{"echo 'test with | pipe'", "echo 'test & ampersand'\n\necho new line"})
		require.NoError(t, err)
		content, err := os.ReadFile(args[1])
		require.NoError(t, err)
		assert.EqualValues(t, `@SET PROMPT=$
@echo + 'echo '"'"'test with | pipe'"'"''
@echo 'test with | pipe'
@IF NOT %ERRORLEVEL% == 0 exit %ERRORLEVEL%
@echo + 'echo '"'"'test & ampersand'"'"'

echo new line'
@echo 'test & ampersand'

echo new line
@IF NOT %ERRORLEVEL% == 0 exit %ERRORLEVEL%
`, string(content))
	})

	t.Run("shell with .exe suffix", func(t *testing.T) {
		args, err := e.genCmdByShell("bash.exe", []string{"echo test"})
		require.NoError(t, err)
		assert.Len(t, args, 3)
		assert.Equal(t, "-e", args[0])
	})
}

func TestProbeShellIsPosix(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("skipping posix shell tests on non-linux system")
	}

	t.Run("valid posix shells", func(t *testing.T) {
		err := probeShellIsPosix("sh")
		assert.NoError(t, err)
	})

	t.Run("invalid shell", func(t *testing.T) {
		err := probeShellIsPosix("nonexistentshell12345")
		if assert.ErrorIs(t, err, ErrNoPosixShell) {
			assert.Equal(t, "assumed posix shell but test failed: shell 'nonexistentshell12345' returned: exec: \"nonexistentshell12345\": executable file not found in $PATH", err.Error())
		}
	})

	t.Run("non-posix shell", func(t *testing.T) {
		// nologin won't understand posix syntax
		err := probeShellIsPosix("true")
		if assert.ErrorIs(t, err, ErrNoPosixShell) {
			assert.Equal(t, "assumed posix shell but test failed: shell 'true' returned unexpected output: ''", err.Error())
		}
	})
}
