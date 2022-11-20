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

	"github.com/alessio/shellescape"
)

// generateScriptPosix is a helper function that generates a step script
// for a linux container using the given
func generateScriptPosix(commands []string) string {
	var buf bytes.Buffer
	for _, command := range commands {
		buf.WriteString(fmt.Sprintf(
			traceScript,
			shellescape.Quote(command),
			command,
		))
	}
	script := fmt.Sprintf(
		setupScript,
		buf.String(),
	)
	return script
}

// setupScript is a helper script this is added to the step script to ensure
// a minimum set of environment variables are set correctly.
const setupScript = `
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
%s
`

// traceScript is a helper script that is added to the step script
// to trace a command.
const traceScript = `
echo + %s
%s
`
