// Copyright 2024 Woodpecker Authors
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

const (
	windowsScriptBase64 = "CiRFcnJvckFjdGlvblByZWZlcmVuY2UgPSAnU3RvcCc7CmlmIChbRW52aXJvbm1lbnRdOjpHZXRFbnZpcm9ubWVudFZhcmlhYmxlKCdDSV9XT1JLU1BBQ0UnKSkgeyBpZiAoLW5vdCAoVGVzdC1QYXRoICIkZW52OkNJX1dPUktTUEFDRSIpKSB7IE5ldy1JdGVtIC1QYXRoICIkZW52OkNJX1dPUktTUEFDRSIgLUl0ZW1UeXBlIERpcmVjdG9yeSAtRm9yY2UgfX07CmlmICgtbm90IFtFbnZpcm9ubWVudF06OkdldEVudmlyb25tZW50VmFyaWFibGUoJ0hPTUUnKSkgeyBbRW52aXJvbm1lbnRdOjpTZXRFbnZpcm9ubWVudFZhcmlhYmxlKCdIT01FJywgJ2M6XHJvb3QnKSB9OwppZiAoLW5vdCAoVGVzdC1QYXRoICIkZW52OkhPTUUiKSkgeyBOZXctSXRlbSAtUGF0aCAiJGVudjpIT01FIiAtSXRlbVR5cGUgRGlyZWN0b3J5IC1Gb3JjZSB9OwppZiAoJEVudjpDSV9ORVRSQ19NQUNISU5FKSB7CiRuZXRyYz1bc3RyaW5nXTo6Rm9ybWF0KCJ7MH1cX25ldHJjIiwkRW52OkhPTUUpOwoibWFjaGluZSAkRW52OkNJX05FVFJDX01BQ0hJTkUiID4+ICRuZXRyYzsKImxvZ2luICRFbnY6Q0lfTkVUUkNfVVNFUk5BTUUiID4+ICRuZXRyYzsKInBhc3N3b3JkICRFbnY6Q0lfTkVUUkNfUEFTU1dPUkQiID4+ICRuZXRyYzsKfTsKW0Vudmlyb25tZW50XTo6U2V0RW52aXJvbm1lbnRWYXJpYWJsZSgiQ0lfTkVUUkNfUEFTU1dPUkQiLCRudWxsKTsKW0Vudmlyb25tZW50XTo6U2V0RW52aXJvbm1lbnRWYXJpYWJsZSgiQ0lfU0NSSVBUIiwkbnVsbCk7CmlmIChbRW52aXJvbm1lbnRdOjpHZXRFbnZpcm9ubWVudFZhcmlhYmxlKCdDSV9XT1JLU1BBQ0UnKSkgeyBjZCAiJGVudjpDSV9XT1JLU1BBQ0UiIH07CgpXcml0ZS1PdXRwdXQgKCcrICJlY2hvIGhlbGxvIHdvcmxkIicpOwomIGVjaG8gaGVsbG8gd29ybGQ7IGlmICgkTEFTVEVYSVRDT0RFIC1uZSAwKSB7ZXhpdCAkTEFTVEVYSVRDT0RFfQoK"
	posixScriptBase64   = "CmlmIFsgLW4gIiRDSV9ORVRSQ19NQUNISU5FIiBdOyB0aGVuCmNhdCA8PEVPRiA+ICRIT01FLy5uZXRyYwptYWNoaW5lICRDSV9ORVRSQ19NQUNISU5FCmxvZ2luICRDSV9ORVRSQ19VU0VSTkFNRQpwYXNzd29yZCAkQ0lfTkVUUkNfUEFTU1dPUkQKRU9GCmNobW9kIDA2MDAgJEhPTUUvLm5ldHJjCmZpCnVuc2V0IENJX05FVFJDX1VTRVJOQU1FCnVuc2V0IENJX05FVFJDX1BBU1NXT1JECnVuc2V0IENJX1NDUklQVApta2RpciAtcCAiJENJX1dPUktTUEFDRSIKY2QgIiRDSV9XT1JLU1BBQ0UiCgplY2hvICsgJ2VjaG8gaGVsbG8gd29ybGQnCmVjaG8gaGVsbG8gd29ybGQK"
)

func TestGenerateContainerConf(t *testing.T) {
	gotEnv, gotEntry := GenerateContainerConf([]string{"echo hello world"}, "windows")
	assert.Equal(t, windowsScriptBase64, gotEnv["CI_SCRIPT"])
	assert.Equal(t, "powershell.exe", gotEnv["SHELL"])
	assert.Equal(t, []string{"powershell", "-noprofile", "-noninteractive", "-command", "[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($Env:CI_SCRIPT)) | iex"}, gotEntry)
	gotEnv, gotEntry = GenerateContainerConf([]string{"echo hello world"}, "linux")
	assert.Equal(t, posixScriptBase64, gotEnv["CI_SCRIPT"])
	assert.Equal(t, "/bin/sh", gotEnv["SHELL"])
	assert.Equal(t, []string{"/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"}, gotEntry)
}
