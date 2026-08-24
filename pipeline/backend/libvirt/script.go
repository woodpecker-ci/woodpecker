// Copyright 2026 Julian Ospald
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

package libvirt

import (
	"maps"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func GenerateSSHConf(step *backend_types.Step, osType string) (stdin string, entry []string, err error) {
	env := make(map[string]string)
	maps.Copy(env, step.Environment)
	var script string
	if osType == "windows" {
		env["SHELL"] = "powershell.exe"
		env["CI_WRITE_PID"] = "yes"

		// The libvirt backend does not expose the CI_SCRIPT env var.
		// Because we inject the `export CI_FOO=bar` expressions into the script itself,
		// this would pose two problems:
		//
		// 1. CI_SCRIPT is now self-referencing
		// 2. CI_SCRIPT would contain the secrets
		script = generateScriptWindows(step.Commands, env, step.WorkingDir, step.UUID, true)

		entry = []string{"powershell", "-noprofile", "-noninteractive", "-command", "-"}
	} else {
		env["SHELL"] = "/bin/sh"
		env["CI_WRITE_PID"] = "yes"

		// same comment as for windows applies
		script = generateScriptPosix(step.Commands, env, step.WorkingDir, step.UUID, true)

		// cspell:disable-next-line
		entry = []string{"/bin/sh", "-e"}
	}

	return script, entry, nil
}
