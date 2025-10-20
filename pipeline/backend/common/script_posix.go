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
	"text/template"

	"al.essio.dev/pkg/shellescape"
)

// generateScriptPosix is a helper function that generates a step script
// for a linux container using the given.
func generateScriptPosix(commands []string, workDir string) string {
	var buf bytes.Buffer

	if err := setupScriptTmpl.Execute(&buf, map[string]string{
		"WorkDir": workDir,
	}); err != nil {
		// should never happen but well we have an error to trance
		return fmt.Sprintf("echo 'failed to generate posix script from commands: %s'; exit 1", err.Error())
	}

	for _, command := range commands {
		buf.WriteString(fmt.Sprintf(
			traceScript,
			shellescape.Quote(command),
			command,
		))
	}

	return buf.String()
}

// setupScriptProto is a helper script this is added to the step script to ensure
// a minimum set of environment variables are set correctly.
const setupScriptProto = `
if [ -n "$CI_NETRC_MACHINE" ]; then
cat <<EOF > $HOME/.netrc
machine $CI_NETRC_MACHINE
login $CI_NETRC_USERNAME
password $CI_NETRC_PASSWORD
EOF
chmod 0600 $HOME/.netrc
fi
unset CI_NETRC_USERNAME
unset CI_NETRC_PASSWORD
unset CI_SCRIPT
mkdir -p "{{.WorkDir}}"
cd "{{.WorkDir}}"
`

var setupScriptTmpl, _ = template.New("").Parse(setupScriptProto)

// traceScript is a helper script that is added to the step script
// to trace a command.
const traceScript = `
echo + %s
%s
`
