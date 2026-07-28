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
	"github.com/urfave/cli/v3"
)

type config struct {
	binPath       string
	tempDir       string   // temp dir for workflow workspaces
	isolatedHome  bool     // set HOME/USERPROFILE to isolated dir
	seccompFile   string
	seccompString string
	maxCPU        int
	maxMemory     int
	maxNofile     int
	maxPids       int
	readonly      bool
	procMount     bool   // mount /proc from host (default: off)
	noNewPrivs    bool
	isolateNet    bool
	isolatePid    bool
	isolateIpc    bool
	isolateUts    bool
	isolateUser   bool
	uid           int
	gid           int
	timeLimit     int
	cgroupPidsMax int
	cgroupCpuMs   int
}

func configFromCli(c *cli.Command) config {
	return config{
		binPath:       c.String("backend-nsjail-bin"),
		tempDir:       c.String("backend-nsjail-temp-dir"),
		isolatedHome:  c.Bool("backend-nsjail-isolated-home"),
		seccompFile:   c.String("backend-nsjail-seccomp"),
		seccompString: c.String("backend-nsjail-seccomp-string"),
		maxCPU:        c.Int("backend-nsjail-max-cpu"),
		maxMemory:     c.Int("backend-nsjail-max-memory"),
		maxNofile:     c.Int("backend-nsjail-max-nofile"),
		maxPids:       c.Int("backend-nsjail-max-pids"),
		readonly:      c.Bool("backend-nsjail-readonly"),
		procMount:     c.Bool("backend-nsjail-proc-mount"),
		noNewPrivs:    c.Bool("backend-nsjail-no-new-privs"),
		isolateNet:    c.Bool("backend-nsjail-isolate-net"),
		isolatePid:    c.Bool("backend-nsjail-isolate-pid"),
		isolateIpc:    c.Bool("backend-nsjail-isolate-ipc"),
		isolateUts:    c.Bool("backend-nsjail-isolate-uts"),
		isolateUser:   c.Bool("backend-nsjail-isolate-user"),
		uid:           c.Int("backend-nsjail-uid"),
		gid:           c.Int("backend-nsjail-gid"),
		timeLimit:     c.Int("backend-nsjail-time-limit"),
		cgroupPidsMax: c.Int("backend-nsjail-cgroup-pids-max"),
		cgroupCpuMs:   c.Int("backend-nsjail-cgroup-cpu-ms"),
	}
}
