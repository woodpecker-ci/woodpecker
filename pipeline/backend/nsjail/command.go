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

// cSpell:ignore ERRORLEVEL

package nsjail

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"al.essio.dev/pkg/shellescape"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

// defaultSeccompProfile returns an architecture-appropriate default seccomp policy.
// Kafel syscall names differ between architectures (e.g., arm64 lacks open, access, arch_prctl).
// Users can override via --backend-nsjail-seccomp / --backend-nsjail-seccomp-string.
func defaultSeccompProfile() string {
	if runtime.GOARCH == "arm64" {
		return arm64SeccompProfile
	}
	return amd64SeccompProfile
}

const amd64SeccompProfile = `
POLICY ci_default {
    ALLOW {
        read, write, open, openat, close,
        mmap, munmap, brk, mprotect,
        fstat, fstatat, newfstatat, lstat, stat,
        access, faccessat, faccessat2,
        lseek, pread64, pwrite64,
        readv, writev, preadv, pwritev,
        unlink, unlinkat, rmdir, mkdir, mkdirat,
        rename, renameat, renameat2,
        link, linkat, symlink, symlinkat,
        readlink, readlinkat, getdents, getdents64,
        chdir, fchdir, fchmod, fchmodat,
        chown, fchown, lchown, fchownat,
        copy_file_range, splice, sendfile,
        fallocate, ftruncate, truncate,
        utimensat, futimesat,
        clone, clone3, fork, vfork,
        execve, execveat, exit_group,
        getpid, getppid, gettid,
        getuid, getgid, geteuid, getegid,
        getresuid, getresgid,
        setuid, setgid, setresuid, setresgid,
        set_tid_address, set_robust_list,
        wait4, waitid,
        mremap, msync, mincore, madvise,
        shmget, shmat, shmdt, shmctl,
        rt_sigreturn, rt_sigaction, rt_sigprocmask,
        rt_sigpending, rt_sigsuspend, sigaltstack,
        kill, tkill, tgkill,
        socket, connect, bind, listen, accept, accept4,
        sendto, recvfrom, sendmsg, recvmsg,
        sendmmsg, recvmmsg,
        setsockopt, getsockopt,
        shutdown, getsockname, getpeername,
        nanosleep, clock_nanosleep,
        sched_yield,
        clock_gettime, clock_settime,
        gettimeofday, settimeofday, time,
        futex, futex_time64,
        arch_prctl, prctl,
        sysinfo, uname,
        pipe, pipe2,
        eventfd, eventfd2, epoll_create,
        epoll_create1, epoll_ctl, epoll_wait,
        poll, ppoll, select, pselect6,
        ioctl,
        getrandom,
        capget, capset,
        statfs, statfs64, fstatfs, fstatfs64,
        mount, umount, umount2,
        personality
    }
}
USE ci_default DEFAULT KILL
`

// arm64SeccompProfile uses syscall names from the ARM64 generic syscall ABI.
// ARM64 uses a simplified ABI that lacks many legacy syscalls (open, access, dup2, pipe, etc.).
const arm64SeccompProfile = `
POLICY ci_default {
    ALLOW {
        read, write, close,
        mmap, munmap, mprotect, brk, madvise,
        openat, newfstatat, statx, faccessat,
        lseek, pread64, pwrite64,
        readv, writev, preadv, pwritev,
        unlinkat, mkdirat, renameat, symlinkat, readlinkat,
        getdents64,
        chdir, fchdir, fchmodat, fchownat,
        copy_file_range,
        clone, clone3,
        execve, execveat, exit_group,
        getpid, getppid, gettid,
        getuid, getgid, geteuid, getegid,
        setuid, setgid,
        set_tid_address, set_robust_list,
        waitid,
        mremap, msync, mincore,
        rt_sigreturn, rt_sigaction, rt_sigprocmask,
        rt_sigpending, rt_sigsuspend, sigaltstack,
        kill, tkill, tgkill,
        socket, connect, bind, listen, accept4,
        sendto, recvfrom, sendmsg, recvmsg,
        sendmmsg, recvmmsg,
        setsockopt, getsockopt,
        shutdown, getsockname, getpeername,
        nanosleep, clock_nanosleep,
        sched_yield,
        clock_gettime, clock_settime,
        gettimeofday, settimeofday,
        futex,
        prctl,
        pipe2, eventfd2, epoll_create1, epoll_ctl, epoll_pwait,
        ppoll, pselect6,
        ioctl, fcntl,
        getrandom,
        capget, capset,
        statfs, fstatfs,
        getcwd, dup, dup3, umask, fchmod,
        fchown, linkat, renameat2, truncate, ftruncate,
        setpgid, getpgid, setsid, getsid,
        flock, fsync, fdatasync,
        symlinkat, readlinkat,
        faccessat2
    }
}
USE ci_default DEFAULT KILL
`

// buildNsJailArgs converts a Woodpecker Step into nsjail CLI arguments.
// All security hardening is enabled by default; config can override.
func (e *nsjail) buildNsJailArgs(step *types.Step, state *workflowState, env []string) []string {
	var args []string

	// === Mode: ONCE (execute once then exit) ===
	args = append(args, "-Mo")

	// === Chroot ===
	args = append(args, "--chroot", "/")

	// === Resource limits ===
	if e.cfg.maxCPU > 0 {
		args = append(args, "--rlimit_cpu", strconv.Itoa(e.cfg.maxCPU))
	} else {
		args = append(args, "--rlimit_cpu", "300")
	}
	if e.cfg.maxMemory > 0 {
		args = append(args, "--rlimit_as", strconv.Itoa(e.cfg.maxMemory))
	} else {
		args = append(args, "--rlimit_as", "512")
	}
	if e.cfg.maxNofile > 0 {
		args = append(args, "--rlimit_nofile", strconv.Itoa(e.cfg.maxNofile))
	} else {
		args = append(args, "--rlimit_nofile", "64")
	}
	if e.cfg.maxPids > 0 {
		args = append(args, "--rlimit_nproc", strconv.Itoa(e.cfg.maxPids))
	} else {
		args = append(args, "--rlimit_nproc", "64")
	}

	// === cgroup resource limits ===
	// detect_cgroupv2: auto-detect cgroup v2 if available (needed on modern Linux)
	args = append(args, "--detect_cgroupv2")
	if e.cfg.cgroupPidsMax > 0 {
		args = append(args, "--cgroup_pids_max", strconv.Itoa(e.cfg.cgroupPidsMax))
	} else {
		args = append(args, "--cgroup_pids_max", "64")
	}
	if e.cfg.cgroupCpuMs > 0 {
		args = append(args, "--cgroup_cpu_ms_per_sec", strconv.Itoa(e.cfg.cgroupCpuMs))
	} else {
		args = append(args, "--cgroup_cpu_ms_per_sec", "800")
	}

	// === Time limit ===
	if e.cfg.timeLimit > 0 {
		args = append(args, "--time_limit", strconv.Itoa(e.cfg.timeLimit))
	} else {
		args = append(args, "--time_limit", "600")
	}

	// === Seccomp policy ===
	// seccompString="off" disables seccomp entirely (omit --seccomp_string flag)
	if e.cfg.seccompString != "" && e.cfg.seccompString != "off" {
		args = append(args, "--seccomp_string", e.cfg.seccompString)
	} else if e.cfg.seccompFile != "" {
		args = append(args, "--seccomp", e.cfg.seccompFile)
	} else if e.cfg.seccompString != "off" {
		// Default seccomp policy when nothing is configured
		args = append(args, "--seccomp_string", defaultSeccompProfile())
	}

	// === User/Group mapping ===
	// Default to nobody (UID/GID 65534), overridable via --backend-nsjail-uid / --backend-nsjail-gid
	jailUID := 65534
	if e.cfg.uid > 0 {
		jailUID = e.cfg.uid
	}
	jailGID := 65534
	if e.cfg.gid > 0 {
		jailGID = e.cfg.gid
	}
	args = append(args, "--user", strconv.Itoa(jailUID))
	args = append(args, "--group", strconv.Itoa(jailGID))
	args = append(args, "--uid_mapping", fmt.Sprintf("0:%d:1", jailUID))
	args = append(args, "--gid_mapping", fmt.Sprintf("0:%d:1", jailGID))

	// === Filesystem ===
	// When readonly=true mount as read-only, otherwise chroot / allows read-write access by default
	if e.cfg.readonly {
		args = append(args, "--ro", state.workspaceDir)
		args = append(args, "--ro", state.homeDir)
	}

	// === Namespace isolation ===
	if !e.cfg.isolateNet {
		args = append(args, "--disable_clone_newnet")
	}
	if !e.cfg.isolatePid {
		args = append(args, "--disable_clone_newpid")
	}
	if !e.cfg.isolateIpc {
		args = append(args, "--disable_clone_newipc")
	}
	if !e.cfg.isolateUts {
		args = append(args, "--disable_clone_newuts")
	}
	if !e.cfg.isolateUser {
		args = append(args, "--disable_clone_newuser")
	}

	// === Security hardening ===
	if e.cfg.noNewPrivs {
		args = append(args, "--no_new_privs")
	}

	// === /proc mount (default: off) ===
	if e.cfg.procMount {
		args = append(args, "--bindmount", "/proc:/proc")
	}

	// === Environment variables ===
	// nsjail does NOT inherit parent process env — each --env flag is explicit
	for _, envVar := range env {
		args = append(args, "--env", envVar)
	}

	// === Hostname isolation ===
	args = append(args, "--hostname", "nsjail")

	// === Separator: end of nsjail args ===
	args = append(args, "--")

	return args
}

// execCommands executes a commands step inside the nsjail sandbox.
func (e *nsjail) execCommands(ctx context.Context, step *types.Step, state *workflowState, env []string) error {
	// Look up the full path of the shell binary (e.g., "bash" → "/bin/bash")
	// nsjail's execve() inside the jail needs an absolute path
	shellPath, err := exec.LookPath(step.Image)
	if err != nil {
		return fmt.Errorf("shell %q not found: %w", step.Image, err)
	}

	// Build nsjail base arguments (ending with "--")
	nsjailArgs := e.buildNsJailArgs(step, state, env)

	// Get shell arguments via genCmdByShell (different shell types use different argument formats)
	shellArgs, err := e.genCmdByShell(step.Image, step.Commands, state.baseDir)
	if err != nil {
		return fmt.Errorf("could not convert commands into args: %w", err)
	}

	// Append shell binary and its arguments after "--"
	// nsjail expects the first arg after "--" to be the executable path
	nsjailArgs = append(nsjailArgs, shellPath)
	nsjailArgs = append(nsjailArgs, shellArgs...)

	cmd := newNsJailCmd(ctx, e.cfg.binPath, nsjailArgs...)
	cmd.Env = env
	cmd.Dir = state.workspaceDir

	reader, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	cmd.Stderr = cmd.Stdout

	state.stepState.Store(step.UUID, &stepState{cmd: cmd, output: reader})
	return cmd.Start()
}

// genCmdByShell generates shell execution arguments based on shell type (step.Image).
// Supports: sh/bash/zsh (POSIX), cmd, fish, nu, powershell/pwsh.
func (e *nsjail) genCmdByShell(shell string, cmdList []string, baseDir string) (args []string, err error) {
	if len(cmdList) == 0 {
		return nil, ErrNoCmdSet
	}

	script := ""
	for _, cmd := range cmdList {
		script += fmt.Sprintf("echo %s\n%s\n", strings.TrimSpace(shellescape.Quote("+ "+cmd)), cmd)
	}
	script = strings.TrimSpace(script)

	shell = strings.TrimSuffix(strings.ToLower(shell), ".exe")
	switch shell {
	default:
		// assume posix shell
		if err := probeShellIsPosix(shell); err != nil {
			return nil, err
		}
		fallthrough
	case "sh", "bash", "zsh":
		return []string{"-e", "-c", script}, nil
	case "":
		return nil, ErrNoShellSet
	case "cmd":
		script := "@SET PROMPT=$\n"
		for _, cmd := range cmdList {
			quotedCmd := strings.TrimSpace(shellescape.Quote(cmd))
			quotedCmd = strings.ReplaceAll(quotedCmd, "\n", "\\n")
			quotedCmd = strings.ReplaceAll(quotedCmd, "&", "\\AND")
			quotedCmd = strings.ReplaceAll(quotedCmd, "|", "\\OR")
			script += fmt.Sprintf("@echo + %s\n", quotedCmd)
			script += fmt.Sprintf("@%s\n", cmd)
			script += "@IF NOT %ERRORLEVEL% == 0 exit %ERRORLEVEL%\n"
		}
		cmdFile, err := os.CreateTemp(baseDir, "*.cmd")
		if err != nil {
			return nil, err
		}
		defer cmdFile.Close()
		if _, err := cmdFile.WriteString(script); err != nil {
			return nil, err
		}
		return []string{"/c", cmdFile.Name()}, nil
	case "fish":
		script := ""
		for _, cmd := range cmdList {
			script += fmt.Sprintf("echo %s\n%s || exit $status\n", strings.TrimSpace(shellescape.Quote("+ "+cmd)), cmd)
		}
		return []string{"-c", script}, nil
	case "nu":
		return []string{"--commands", script}, nil
	case "powershell", "pwsh":
		return []string{"-noprofile", "-noninteractive", "-c", "$ErrorActionPreference = \"Stop\"; " + script}, nil
	}
}

// probeShellIsPosix checks if the given shell is POSIX-compatible.
func probeShellIsPosix(shell string) error {
	script := `x=1 && [ "$x" = "1" ] && command -v test >/dev/null && printf ok`

	cmd := exec.Command(shell, "-c", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &ErrNoPosixShell{Shell: shell, Err: err}
	}

	if strings.TrimSpace(string(output)) != "ok" {
		return &ErrNoPosixShell{Shell: shell, Err: fmt.Errorf("unexpected output: %q", string(output))}
	}

	return nil
}

// buildNsJailArgsForClone builds nsjail CLI arguments for clone steps (with bindmount for plugin-git).
// The returned args end with "--", the actual command is appended by execClone.
func (e *nsjail) buildNsJailArgsForClone(step *types.Step, state *workflowState, env []string, pluginGitPath, netrcPath string) []string {
	args := e.buildNsJailArgs(step, state, env)

	// Find "--" separator and truncate to it (keep "--")
	sepIdx := len(args) - 1
	for i := len(args) - 1; i >= 0; i-- {
		if args[i] == "--" {
			sepIdx = i
			break
		}
	}
	args = args[:sepIdx+1] // keep "--"

	// Insert clone-specific bindmount arguments before "--"
	var extraArgs []string
	if pluginGitPath != "" {
		extraArgs = append(extraArgs, "--bindmount_ro", pluginGitPath+":/bin/plugin-git")
	}
	if netrcPath != "" {
		netrcInside := filepath.Join(state.homeDir, ".netrc")
		extraArgs = append(extraArgs, "--bindmount", netrcPath+":"+netrcInside)
	}

	// Insert extra args before "--"
	args = append(args[:sepIdx], append(extraArgs, args[sepIdx:]...)...)

	return args // ends with "--", command appended by execClone
}

// newNsJailCmd creates an nsjail command with process group support for batch kill.
func newNsJailCmd(ctx context.Context, binPath string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, binPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	return cmd
}
