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
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

func generateScriptWindows(commands []string, workDir string) string {
	var buf bytes.Buffer

	if err := setupScriptTmpl.Execute(&buf, map[string]string{
		"WorkDir": workDir,
	}); err != nil {
		// should never happen but well we have an error to trance
		return fmt.Sprintf("echo 'failed to generate posix script from commands: %s'; exit 1", err.Error())
	}

	for _, command := range commands {
		escaped := fmt.Sprintf("%q", command)
		escaped = strings.ReplaceAll(escaped, "$", `\$`)
		buf.WriteString(fmt.Sprintf(
			traceScriptWin,
			escaped,
			command,
		))
	}

	return buf.String()
}

const setupScriptWinProto = `
$ErrorActionPreference = 'Stop';
if ([Environment]::GetEnvironmentVariable('CI_WORKSPACE')) { if (-not (Test-Path "{{.WorkDir}}")) { New-Item -Path "{{.WorkDir}}" -ItemType Directory -Force }};
if (-not [Environment]::GetEnvironmentVariable('HOME')) { [Environment]::SetEnvironmentVariable('HOME', 'c:\root') };
if (-not (Test-Path "$env:HOME")) { New-Item -Path "$env:HOME" -ItemType Directory -Force };
if ($Env:CI_NETRC_MACHINE) {
$netrc=[string]::Format("{0}\_netrc",$Env:HOME);
"machine $Env:CI_NETRC_MACHINE" >> $netrc;
"login $Env:CI_NETRC_USERNAME" >> $netrc;
"password $Env:CI_NETRC_PASSWORD" >> $netrc;
};
[Environment]::SetEnvironmentVariable("CI_NETRC_PASSWORD",$null);
[Environment]::SetEnvironmentVariable("CI_SCRIPT",$null);
if ([Environment]::GetEnvironmentVariable('CI_WORKSPACE')) { cd "{{.WorkDir}}" };
`

var setupScriptWinTmpl, _ = template.New("").Parse(setupScriptWinProto)

// traceScript is a helper script that is added to the step script
// to trace a command.
const traceScriptWin = `
Write-Output ('+ %s');
& %s; if ($LASTEXITCODE -ne 0) {exit $LASTEXITCODE}
`
