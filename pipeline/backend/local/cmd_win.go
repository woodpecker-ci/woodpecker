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

//go:build windows

package local

import (
	"context"
	"os/exec"
	"strconv"
)

func newCmd(ctx context.Context, binary string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, binary, args...)

	// Non perfect workaround till std exec supports JOB_OBJECT
	// https://github.com/woodpecker-ci/woodpecker/issues/6717 & https://github.com/golang/go/issues/79927
	cmd.Cancel = func() error {
		return exec.Command("taskkill", "/F", "/T", "/PID",
			strconv.Itoa(cmd.Process.Pid)).Run()
	}

	return cmd
}
