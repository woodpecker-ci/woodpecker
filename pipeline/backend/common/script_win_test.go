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

func TestGenerateScriptWin(t *testing.T) {
	testdata := []struct {
		from []string
		want string
	}{
		{
			from: []string{"echo %PATH%", "go build", "go test"},
			want: `
$ErrorActionPreference = 'Stop';
&cmd /c "mkdir c:\root";
if ($Env:CI_NETRC_MACHINE) {
$netrc=[string]::Format("{0}\_netrc",$Env:HOME);
"machine $Env:CI_NETRC_MACHINE" >> $netrc;
"login $Env:CI_NETRC_USERNAME" >> $netrc;
"password $Env:CI_NETRC_PASSWORD" >> $netrc;
};
[Environment]::SetEnvironmentVariable("CI_NETRC_PASSWORD",$null);
[Environment]::SetEnvironmentVariable("CI_SCRIPT",$null);

Write-Output ('+ "echo %PATH%"');
& echo %PATH%; if ($LASTEXITCODE -ne 0) {exit $LASTEXITCODE}

Write-Output ('+ "go build"');
& go build; if ($LASTEXITCODE -ne 0) {exit $LASTEXITCODE}

Write-Output ('+ "go test"');
& go test; if ($LASTEXITCODE -ne 0) {exit $LASTEXITCODE}

`,
		},
	}
	for _, test := range testdata {
		script := generateScriptWindows(test.from)
		assert.EqualValues(t, test.want, script, "Want encoded script for %s", test.from)
	}
}
