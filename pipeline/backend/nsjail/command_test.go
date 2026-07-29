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

package nsjail

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func TestGenCmdByShell(t *testing.T) {
	e := &nsjail{}

	t.Run("error cases", func(t *testing.T) {
		args, err := e.genCmdByShell("", []string{"echo hi"}, t.TempDir())
		assert.Nil(t, args)
		assert.ErrorIs(t, err, ErrNoShellSet)

		args, err = e.genCmdByShell("sh", []string{}, t.TempDir())
		assert.Nil(t, args)
		assert.ErrorIs(t, err, ErrNoCmdSet)
	})

	t.Run("unix shells", func(t *testing.T) {
		args, err := e.genCmdByShell("sh", []string{"echo hello", "pwd"}, t.TempDir())
		require.NoError(t, err)
		assert.Len(t, args, 3)
		assert.Equal(t, "-e", args[0])
		assert.Equal(t, "-c", args[1])
		assert.Contains(t, args[2], "echo hello")
		assert.Contains(t, args[2], "pwd")

		args, err = e.genCmdByShell("bash", []string{"ls -la"}, t.TempDir())
		require.NoError(t, err)
		assert.Len(t, args, 3)
		assert.Equal(t, "-e", args[0])

		args, err = e.genCmdByShell("zsh", []string{"echo test"}, t.TempDir())
		require.NoError(t, err)
		assert.Len(t, args, 3)
		assert.Equal(t, "-e", args[0])
	})

	t.Run("windows shells", func(t *testing.T) {
		args, err := e.genCmdByShell("cmd.exe", []string{"echo hi"}, t.TempDir())
		require.NoError(t, err)
		require.Len(t, args, 2)
		assert.Equal(t, "/c", args[0])
		assert.True(t, strings.HasSuffix(args[1], ".cmd"))

		args, err = e.genCmdByShell("powershell", []string{"Write-Host 'test'"}, t.TempDir())
		require.NoError(t, err)
		assert.Len(t, args, 4)
		assert.Equal(t, "-noprofile", args[0])
		assert.Equal(t, "-noninteractive", args[1])
		assert.Equal(t, "-c", args[2])
	})

	t.Run("fish shell", func(t *testing.T) {
		args, err := e.genCmdByShell("fish", []string{"echo test"}, t.TempDir())
		require.NoError(t, err)
		assert.Len(t, args, 2)
		assert.Equal(t, "-c", args[0])
		assert.Contains(t, args[1], "|| exit $status")
	})

	t.Run("nu shell", func(t *testing.T) {
		args, err := e.genCmdByShell("nu", []string{"echo test"}, t.TempDir())
		require.NoError(t, err)
		assert.Len(t, args, 2)
		assert.Equal(t, "--commands", args[0])
	})

	t.Run("cmd temp file content", func(t *testing.T) {
		args, err := e.genCmdByShell("cmd", []string{"echo hello"}, t.TempDir())
		require.NoError(t, err)
		content, err := os.ReadFile(args[1])
		require.NoError(t, err)
		assert.Contains(t, string(content), "@echo + 'echo hello'")
		assert.Contains(t, string(content), "@echo hello")
	})
}

func TestBuildNsJailArgs(t *testing.T) {
	t.Run("custom values", func(t *testing.T) {
		e := &nsjail{}
		e.cfg.timeLimit = 30
		e.cfg.maxCPU = 60

		step := &types.Step{UUID: "test-step"}
		state := &workflowState{
			workspaceDir: "/tmp/workspace",
			homeDir:      "/tmp/home",
		}

		args := e.buildNsJailArgs(step, state, nil)

		// Check essential args
		assert.Contains(t, args, "-Mo")
		assert.Contains(t, args, "--chroot")
		assert.Contains(t, args, "/")
		assert.Contains(t, args, "--time_limit")
		assert.Contains(t, args, "30")
		assert.Contains(t, args, "--rlimit_cpu")
		assert.Contains(t, args, "60")
		assert.Contains(t, args, "--hostname")
		assert.Contains(t, args, "nsjail")
		assert.Contains(t, args, "--user")
		assert.Contains(t, args, "65534")
		assert.Contains(t, args, "--group")
		assert.Contains(t, args, "65534")
		assert.Contains(t, args, "--seccomp_string")
		assert.Contains(t, args, "--")

		// Verify separator position (last element before command)
		sepIdx := -1
		for i, a := range args {
			if a == "--" {
				sepIdx = i
				break
			}
		}
		assert.Greater(t, sepIdx, 0, "-- separator should exist")
	})

	t.Run("default values", func(t *testing.T) {
		e := &nsjail{}

		args := e.buildNsJailArgs(&types.Step{}, &workflowState{}, nil)

		// Default resource limits
		assert.Contains(t, args, "--rlimit_cpu")
		assert.Contains(t, args, "300")
		assert.Contains(t, args, "--rlimit_as")
		assert.Contains(t, args, "512")
		assert.Contains(t, args, "--rlimit_nofile")
		assert.Contains(t, args, "64")
		assert.Contains(t, args, "--rlimit_nproc")
		assert.Contains(t, args, "64")
		assert.Contains(t, args, "--time_limit")
		assert.Contains(t, args, "600")

		// Default cgroup limits
		assert.Contains(t, args, "--cgroup_pids_max")
		assert.Contains(t, args, "64")
		assert.Contains(t, args, "--cgroup_cpu_ms_per_sec")
		assert.Contains(t, args, "800")

		// Default seccomp (built-in policy)
		assert.Contains(t, args, "--seccomp_string")
	})
}

func TestBuildNsJailArgsCustomUID(t *testing.T) {
	e := &nsjail{}
	e.cfg.uid = 1000
	e.cfg.gid = 1000
	e.cfg.noNewPrivs = true
	e.cfg.readonly = true

	step := &types.Step{UUID: "test"}
	state := &workflowState{
		workspaceDir: "/tmp/ws",
		homeDir:      "/tmp/hm",
	}

	args := e.buildNsJailArgs(step, state, nil)

	assert.Contains(t, args, "--user")
	assert.Contains(t, args, "1000")
	assert.Contains(t, args, "--group")
	assert.Contains(t, args, "1000")
	assert.Contains(t, args, "--no_new_privs")
	assert.Contains(t, args, "--ro")
	assert.Contains(t, args, "/tmp/ws")
	assert.Contains(t, args, "/tmp/hm")
}

func TestBuildNsJailArgsDisableIsolation(t *testing.T) {
	e := &nsjail{}
	e.cfg.isolateNet = false
	e.cfg.isolatePid = false

	args := e.buildNsJailArgs(&types.Step{}, &workflowState{}, nil)

	assert.Contains(t, args, "--disable_clone_newnet")
	assert.Contains(t, args, "--disable_clone_newpid")
}

func TestBuildNsJailArgsSeccompFile(t *testing.T) {
	e := &nsjail{}
	e.cfg.seccompFile = "/etc/nsjail/seccomp.cfg"

	args := e.buildNsJailArgs(&types.Step{}, &workflowState{}, nil)

	assert.Contains(t, args, "--seccomp")
	assert.Contains(t, args, "/etc/nsjail/seccomp.cfg")
	assert.NotContains(t, args, "--seccomp_string")
}

func TestBuildNsJailArgsSeccompString(t *testing.T) {
	e := &nsjail{}
	e.cfg.seccompString = "POLICY my { ALLOW { read } DEFAULT KILL } USE my DEFAULT KILL"

	args := e.buildNsJailArgs(&types.Step{}, &workflowState{}, nil)

	assert.Contains(t, args, "--seccomp_string")
	assert.NotContains(t, args, "--seccomp")

	// Find the seccomp_string value in args
	for i, a := range args {
		if a == "--seccomp_string" && i+1 < len(args) {
			assert.Contains(t, args[i+1], "POLICY my")
			return
		}
	}
	t.Error("--seccomp_string not found in args")
}

func TestBuildNsJailArgsForClone(t *testing.T) {
	e := &nsjail{}

	step := &types.Step{UUID: "clone-step"}
	state := &workflowState{
		workspaceDir: "/tmp/ws",
		homeDir:      "/tmp/home",
	}

	args := e.buildNsJailArgsForClone(step, state, nil, "/usr/local/bin/plugin-git", "/tmp/home/.netrc")

	assert.Contains(t, args, "--bindmount_ro")
	assert.Contains(t, args, "/usr/local/bin/plugin-git:/bin/plugin-git")
	assert.Contains(t, args, "--bindmount")

	// Find "--" separator
	sepIdx := -1
	for i, a := range args {
		if a == "--" {
			sepIdx = i
			break
		}
	}
	assert.Greater(t, sepIdx, 0, "-- separator should exist")
	// bindmount args should be before "--"
	assert.Contains(t, args[:sepIdx], "--bindmount_ro")
}

func TestDefaultSeccompProfile(t *testing.T) {
	profile := defaultSeccompProfile()
	assert.Contains(t, profile, "POLICY ci_default")
	assert.Contains(t, profile, "ALLOW {")
	assert.Contains(t, profile, "DEFAULT KILL")
}

func TestDefaultSeccompProfileArchAware(t *testing.T) {
	profile := defaultSeccompProfile()

	// Both profiles must have at least these core syscalls
	core := []string{"read,", "write,", "close,", "mmap,", "execve,", "exit_group,"}
	for _, syscall := range core {
		assert.Contains(t, profile, syscall,
			"default seccomp profile should allow syscall: "+syscall)
	}

	// Architecture-specific checks
	// x86_64 profile has arch_prctl, arm64 profile does not
	if runtime.GOARCH == "arm64" {
		assert.Contains(t, profile, "openat,")
		assert.NotContains(t, profile, "arch_prctl")
	} else {
		assert.Contains(t, profile, "open,")
		assert.Contains(t, profile, "arch_prctl")
	}
}
