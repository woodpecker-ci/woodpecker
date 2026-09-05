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
	"bytes"
	"encoding/base64"
	"fmt"
	"maps"
	"text/template"

	backend_types "go.woodpecker-ci.org/woodpecker/v3/pipeline/backend/types"
)

func GenerateSSHConf(step *backend_types.Step, osType string) (stdin string, entry []string, err error) {
	env := make(map[string]string)
	maps.Copy(env, step.Environment)
	var script string
	if osType == "windows" {
		env["SHELL"] = "powershell.exe"
		env["CI_WRITE_PID"] = "yes"

		// We're injecting the env vars directly into the script for
		// ssh, because otherwise we need to rely on 'AcceptEnv *' sshd config,
		// which is not portable. This has the nasty side effect that it's now
		// harder to generate the CI_SCRIPT env var, which would be self-referencing.
		script = generateScriptWindows(step.Commands, env, step.WorkingDir, step.UUID, true)

		// create the full script without the CI_SCRIPT export to break
		// the self-referencing
		scriptWithoutCIScript, err := ReplaceTemplateVars(script, map[string]string{
			"CiScript": "",
		})
		if err != nil {
			return "", nil, err
		}

		// now we have the contents for CI_SCRIPT
		ci_script := `[Environment]::SetEnvironmentVariable("CI_SCRIPT","` + base64.StdEncoding.EncodeToString([]byte(scriptWithoutCIScript)) + `");`
		// ...and can generate the final script
		script, err = ReplaceTemplateVars(script, map[string]string{
			"CiScript": ci_script,
		})
		if err != nil {
			return "", nil, err
		}

		entry = []string{"powershell", "-noprofile", "-noninteractive", "-command", "-"}
	} else {
		env["SHELL"] = "/bin/sh"
		env["CI_WRITE_PID"] = "yes"

		// same comments as for windows apply
		script = generateScriptPosix(step.Commands, env, step.WorkingDir, step.UUID, true)

		scriptWithoutCIScript, err := ReplaceTemplateVars(script, map[string]string{
			"CiScript": "",
		})
		if err != nil {
			return "", nil, err
		}

		ci_script := `export CI_SCRIPT="` + base64.StdEncoding.EncodeToString([]byte(scriptWithoutCIScript)) + `";`
		script, err = ReplaceTemplateVars(script, map[string]string{
			"CiScript": ci_script,
		})
		if err != nil {
			return "", nil, err
		}

		// cspell:disable-next-line
		entry = []string{"/bin/sh", "-e"}
	}

	return script, entry, nil
}

func ReplaceTemplateVars(script string, tmplVars map[string]string) (string, error) {
	var buf bytes.Buffer
	tmpl, _ := template.New("").Parse(script)
	if err := tmpl.Execute(&buf, tmplVars); err != nil {
		return "", fmt.Errorf("echo 'failed to generate script: %s'; exit 1", err.Error())
	}
	return buf.String(), nil
}
