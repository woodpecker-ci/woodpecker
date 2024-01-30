// Copyright 2023 Woodpecker Authors
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

package local

import (
	"fmt"
	"os"
	"strings"

	"github.com/alessio/shellescape"
)

func (e *local) genCmdByShell(shell string, cmds []string) (args []string, err error) {
	script := ""
	for _, cmd := range cmds {
		script += fmt.Sprintf("echo %s\n%s\n", strings.TrimSpace(shellescape.Quote("+ "+cmd)), cmd)
	}
	script = strings.TrimSpace(script)

	switch strings.TrimSuffix(strings.ToLower(shell), ".exe") {
	case "cmd":
		script := "@SET PROMPT=$\n"
		for _, cmd := range cmds {
			script += fmt.Sprintf("@echo + %s\n", strings.TrimSpace(shellescape.Quote(cmd)))
			script += fmt.Sprintf("@%s\n", cmd)
			script += "@IF NOT %ERRORLEVEL% == 0 exit %ERRORLEVEL%\n"
		}
		cmd, err := os.CreateTemp(e.tempDir, "*.cmd")
		if err != nil {
			return nil, err
		}
		defer cmd.Close()
		if _, err := cmd.WriteString(script); err != nil {
			return nil, err
		}
		return []string{"/c", cmd.Name()}, nil
	case "fish":
		script := ""
		for _, cmd := range cmds {
			script += fmt.Sprintf("echo %s\n%s || exit $status\n", strings.TrimSpace(shellescape.Quote("+ "+cmd)), cmd)
		}
		return []string{"-c", script}, nil
	case "powershell", "pwsh":
		return []string{"-noprofile", "-noninteractive", "-c", "$ErrorActionPreference = \"Stop\"; " + script}, nil
	default:
		// normal posix shells
		return []string{"-e", "-c", script}, nil
	}
}
