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

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateContainerConf(t *testing.T) {
	env, entry, cmd := GenerateContainerConf([]string{"echo ja"}, "linux")
	assert.EqualValues(t, "/root", env["HOME"])
	assert.EqualValues(t, "/bin/sh", env["SHELL"])
	assert.EqualValues(t, []string{"/bin/sh", "-c"}, entry)
	assert.EqualValues(t, []string{"echo $CI_SCRIPT | base64 -d | /bin/sh -e"}, cmd)

	env, entry, cmd = GenerateContainerConf([]string{"echo ja"}, "windows")
	assert.EqualValues(t, "c:\\root", env["HOME"])
	assert.EqualValues(t, "powershell.exe", env["SHELL"])
	assert.EqualValues(t, []string{"powershell", "-noprofile", "-noninteractive", "-command"}, entry)
	assert.EqualValues(t, []string{"[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($Env:CI_SCRIPT)) | iex"}, cmd)
}
