// Copyright 2022 Woodpecker Authors
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

package common

import "runtime"

func GenerateDockerConf(commands []string) (env map[string]string, entry, cmd []string) {
	env = make(map[string]string)
	if runtime.GOOS == "windows" {
		env["CI_SCRIPT"] = generateScriptWindows(commands)
		env["HOME"] = "c:\\root"
		env["SHELL"] = "powershell.exe"
		entry = []string{"powershell", "-noprofile", "-noninteractive", "-command"}
		cmd = []string{"[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($Env:CI_SCRIPT)) | iex"}
	} else {
		env["CI_SCRIPT"] = generateScriptPosix(commands)
		env["HOME"] = "/root"
		env["SHELL"] = "/bin/sh"
		entry = []string{"/bin/sh", "-c"}
		cmd = []string{"echo $CI_SCRIPT | base64 -d | /bin/sh -e"}
	}

	return env, entry, cmd
}

func GenerateScript(commands []string) string {
	if runtime.GOOS == "windows" {
		return generateScriptWindows(commands)
	}
	return generateScriptPosix(commands)
}
