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

import (
	"encoding/base64"

	"golang.org/x/text/encoding/unicode"
)

func GenerateContainerConf(commands []string, osType, workDir string) (env map[string]string, entry []string, err error) {
	env = make(map[string]string)
	if osType == "windows" {
		encoder := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM).NewEncoder()
		utf16leBytes, err := encoder.Bytes([]byte(generateScriptWindows(commands, workDir)))
		if err != nil {
			return nil, nil, err
		}

		env["CI_SCRIPT"] = base64.StdEncoding.EncodeToString(utf16leBytes)
		env["SHELL"] = "powershell.exe"
		// cspell:disable-next-line
		entry = []string{"powershell", "-noprofile", "-noninteractive", "-encodedcommand", env["CI_SCRIPT"]}
	} else {
		env["CI_SCRIPT"] = base64.StdEncoding.EncodeToString([]byte(generateScriptPosix(commands, workDir)))
		env["SHELL"] = "/bin/sh"
		entry = []string{"/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"}
	}

	return env, entry, nil
}
