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
	windowsScriptBase64 = "CiRFcnJvckFjdGlvblByZWZlcmVuY2UgPSAnU3RvcCc7CmlmICgtbm90IChUZXN0LVBhdGggIi93b29kcGVja2VyL3NvbWUiKSkgeyBOZXctSXRlbSAtUGF0aCAiL3dvb2RwZWNrZXIvc29tZSIgLUl0ZW1UeXBlIERpcmVjdG9yeSAtRm9yY2UgfTsKaWYgKC1ub3QgW0Vudmlyb25tZW50XTo6R2V0RW52aXJvbm1lbnRWYXJpYWJsZSgnSE9NRScpKSB7IFtFbnZpcm9ubWVudF06OlNldEVudmlyb25tZW50VmFyaWFibGUoJ0hPTUUnLCAnYzpccm9vdCcpIH07CmlmICgtbm90IChUZXN0LVBhdGggIiRlbnY6SE9NRSIpKSB7IE5ldy1JdGVtIC1QYXRoICIkZW52OkhPTUUiIC1JdGVtVHlwZSBEaXJlY3RvcnkgLUZvcmNlIH07CmlmICgkRW52OkNJX05FVFJDX01BQ0hJTkUpIHsKJG5ldHJjPVtzdHJpbmddOjpGb3JtYXQoInswfVxfbmV0cmMiLCRFbnY6SE9NRSk7CiJtYWNoaW5lICRFbnY6Q0lfTkVUUkNfTUFDSElORSIgPj4gJG5ldHJjOwoibG9naW4gJEVudjpDSV9ORVRSQ19VU0VSTkFNRSIgPj4gJG5ldHJjOwoicGFzc3dvcmQgJEVudjpDSV9ORVRSQ19QQVNTV09SRCIgPj4gJG5ldHJjOwp9OwpbRW52aXJvbm1lbnRdOjpTZXRFbnZpcm9ubWVudFZhcmlhYmxlKCJDSV9ORVRSQ19QQVNTV09SRCIsJG51bGwpOwpbRW52aXJvbm1lbnRdOjpTZXRFbnZpcm9ubWVudFZhcmlhYmxlKCJDSV9TQ1JJUFQiLCRudWxsKTsKY2QgIi93b29kcGVja2VyL3NvbWUiOwoKV3JpdGUtT3V0cHV0ICgnKyAiZWNobyBoZWxsbyB3b3JsZCInKTsKJiBlY2hvIGhlbGxvIHdvcmxkOyBpZiAoJExBU1RFWElUQ09ERSAtbmUgMCkge2V4aXQgJExBU1RFWElUQ09ERX0K"
	posixScriptBase64   = "CnNldCAtZQppZiBbIC1uICIkQ0lfTkVUUkNfTUFDSElORSIgXTsgdGhlbgpjYXQgPDxFT0YgPiAkSE9NRS8ubmV0cmMKbWFjaGluZSAkQ0lfTkVUUkNfTUFDSElORQpsb2dpbiAkQ0lfTkVUUkNfVVNFUk5BTUUKcGFzc3dvcmQgJENJX05FVFJDX1BBU1NXT1JECkVPRgpjaG1vZCAwNjAwICRIT01FLy5uZXRyYwpmaQp1bnNldCBDSV9ORVRSQ19VU0VSTkFNRQp1bnNldCBDSV9ORVRSQ19QQVNTV09SRAp1bnNldCBDSV9TQ1JJUFQKbWtkaXIgLXAgIi93b29kcGVja2VyL3NvbWUiCmNkICIvd29vZHBlY2tlci9zb21lIgoKZWNobyArICdlY2hvIGhlbGxvIHdvcmxkJwplY2hvIGhlbGxvIHdvcmxkCg=="
)

func TestGenerateContainerConf(t *testing.T) {
	gotEnv, gotEntry := GenerateContainerConf([]string{"echo hello world"}, "windows", "/woodpecker/some")
	assert.Equal(t, windowsScriptBase64, gotEnv["CI_SCRIPT"])
	assert.Equal(t, "powershell.exe", gotEnv["SHELL"])
	assert.Equal(t, []string{"powershell", "-noprofile", "-noninteractive", "-command", "[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($Env:CI_SCRIPT)) | iex"}, gotEntry)
	gotEnv, gotEntry = GenerateContainerConf([]string{"echo hello world"}, "linux", "/woodpecker/some")
	assert.Equal(t, posixScriptBase64, gotEnv["CI_SCRIPT"])
	assert.Equal(t, "/bin/sh", gotEnv["SHELL"])
	assert.Equal(t, []string{"/bin/sh", "-c", "echo $CI_SCRIPT | base64 -d | /bin/sh -e"}, gotEntry)
}
