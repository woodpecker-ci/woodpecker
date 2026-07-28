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

	"github.com/urfave/cli/v3"
)

var Flags = []cli.Flag{
	// temp dir
	&cli.StringFlag{
		Name:        "backend-nsjail-temp-dir",
		Sources:     cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_TEMP_DIR"),
		Usage:       "set a different temp dir to clone workflows into",
		DefaultText: "system temporary directory",
		Value:       os.TempDir(),
	},
	&cli.BoolFlag{
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_ISOLATED_HOME"),
		Name:    "backend-nsjail-isolated-home",
		Usage:   "set HOME/USERPROFILE to an isolated directory",
		Value:   true,
	},
	// nsjail 二进制路径
	&cli.StringFlag{
		Name:    "backend-nsjail-bin",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_BIN"),
		Usage:   "path to nsjail binary",
		Value:   "nsjail",
	},
	// Seccomp 策略文件
	&cli.StringFlag{
		Name:    "backend-nsjail-seccomp",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_SECCOMP"),
		Usage:   "path to seccomp policy file (Kafel syntax)",
	},
	// 内联 seccomp 策略（优先级高于文件）
	&cli.StringFlag{
		Name:    "backend-nsjail-seccomp-string",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_SECCOMP_STRING"),
		Usage:   "inline seccomp policy string",
	},
	// 资源限制
	&cli.IntFlag{
		Name:    "backend-nsjail-max-cpu",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_MAX_CPU"),
		Usage:   "CPU time limit in seconds",
	},
	&cli.IntFlag{
		Name:    "backend-nsjail-max-memory",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_MAX_MEMORY"),
		Usage:   "address space limit in MB",
	},
	&cli.IntFlag{
		Name:    "backend-nsjail-max-nofile",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_MAX_NOFILE"),
		Usage:   "max open files",
	},
	&cli.IntFlag{
		Name:    "backend-nsjail-max-pids",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_MAX_PIDS"),
		Usage:   "max number of processes",
	},
	// 挂载
	&cli.BoolFlag{
		Name:    "backend-nsjail-readonly",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_READONLY"),
		Usage:   "mount workspace as read-only",
	},
	// /proc 挂载（默认不挂载，利用 PID namespace 隔离）
	&cli.BoolFlag{
		Name:    "backend-nsjail-proc-mount",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_PROC_MOUNT"),
		Usage:   "mount /proc from host (default: off — prevents host proc leak)",
		Value:   false,
	},
	// 安全
	&cli.BoolFlag{
		Name:    "backend-nsjail-no-new-privs",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_NO_NEW_PRIVS"),
		Usage:   "disallow granting new privileges",
	},
	// Namespace 隔离（默认全部开启）
	&cli.BoolFlag{
		Name:    "backend-nsjail-isolate-net",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_ISOLATE_NET"),
		Usage:   "use new network namespace (default: on)",
		Value:   true,
	},
	&cli.BoolFlag{
		Name:    "backend-nsjail-isolate-pid",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_ISOLATE_PID"),
		Usage:   "use new PID namespace (default: on)",
		Value:   true,
	},
	&cli.BoolFlag{
		Name:    "backend-nsjail-isolate-ipc",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_ISOLATE_IPC"),
		Usage:   "use new IPC namespace (default: on)",
		Value:   true,
	},
	&cli.BoolFlag{
		Name:    "backend-nsjail-isolate-uts",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_ISOLATE_UTS"),
		Usage:   "use new UTS namespace (default: on)",
		Value:   true,
	},
	&cli.BoolFlag{
		Name:    "backend-nsjail-isolate-user",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_ISOLATE_USER"),
		Usage:   "use new user namespace (default: on)",
		Value:   true,
	},
	// 用户/组映射
	&cli.IntFlag{
		Name:    "backend-nsjail-uid",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_UID"),
		Usage:   "UID inside jail",
	},
	&cli.IntFlag{
		Name:    "backend-nsjail-gid",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_GID"),
		Usage:   "GID inside jail",
	},
	// 时间限制
	&cli.IntFlag{
		Name:    "backend-nsjail-time-limit",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_TIME_LIMIT"),
		Usage:   "wall-time limit in seconds",
		Value:   600,
	},
	// cgroup 资源控制
	&cli.IntFlag{
		Name:    "backend-nsjail-cgroup-pids-max",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_CGROUP_PIDS_MAX"),
		Usage:   "cgroup pids max",
	},
	&cli.IntFlag{
		Name:    "backend-nsjail-cgroup-cpu-ms",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_NSJAIL_CGROUP_CPU_MS"),
		Usage:   "cgroup CPU quota (ms per second)",
	},
}
