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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateScriptPosix(t *testing.T) {
	testdata := []struct {
		from []string
		want string
	}{
		{
			from: []string{"echo ${PATH}", "go build", "go test"},
			want: `
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

echo + 'echo ${PATH}'
echo ${PATH}

echo + 'go build'
go build

echo + 'go test'
go test
`,
		},
	}
	for _, test := range testdata {
		script := generateScriptPosix(test.from)
		assert.EqualValues(t, test.want, script, "Want encoded script for %s", test.from)
	}
}
