package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	windowsScriptBase64 = "JEVycm9yQWN0aW9uUHJlZmVyZW5jZSA9ICdTdG9wJzsKJmNtZCAvYyAibWtkaXIgYzpccm9vdCI7CmlmICgkRW52OkNJX05FVFJDX01BQ0hJTkUpIHsKJG5ldHJjPVtzdHJpbmddOjpGb3JtYXQoInswfVxfbmV0cmMiLCRFbnY6SE9NRSk7CiJtYWNoaW5lICRFbnY6Q0lfTkVUUkNfTUFDSElORSIgPj4gJG5ldHJjOwoibG9naW4gJEVudjpDSV9ORVRSQ19VU0VSTkFNRSIgPj4gJG5ldHJjOwoicGFzc3dvcmQgJEVudjpDSV9ORVRSQ19QQVNTV09SRCIgPj4gJG5ldHJjOwp9OwpbRW52aXJvbm1lbnRdOjpTZXRFbnZpcm9ubWVudFZhcmlhYmxlKCJDSV9ORVRSQ19QQVNTV09SRCIsJG51bGwpOwpbRW52aXJvbm1lbnRdOjpTZXRFbnZpcm9ubWVudFZhcmlhYmxlKCJDSV9TQ1JJUFQiLCRudWxsKTsKCldyaXRlLU91dHB1dCAoJysgImVjaG8gaGVsbG8gd29ybGQiJyk7CiYgZWNobyBoZWxsbyB3b3JsZDsgaWYgKCRMQVNURVhJVENPREUgLW5lIDApIHtleGl0ICRMQVNURVhJVENPREV9Cg=="
	posixScriptBase64   = "CmlmIFsgLW4gIiRDSV9ORVRSQ19NQUNISU5FIiBdOyB0aGVuCmNhdCA8PEVPRiA+ICRIT01FLy5uZXRyYwptYWNoaW5lICRDSV9ORVRSQ19NQUNISU5FCmxvZ2luICRDSV9ORVRSQ19VU0VSTkFNRQpwYXNzd29yZCAkQ0lfTkVUUkNfUEFTU1dPUkQKRU9GCmNobW9kIDA2MDAgJEhPTUUvLm5ldHJjCmZpCnVuc2V0IENJX05FVFJDX1VTRVJOQU1FCnVuc2V0IENJX05FVFJDX1BBU1NXT1JECnVuc2V0IENJX1NDUklQVAoKZWNobyArICdlY2hvIGhlbGxvIHdvcmxkJwplY2hvIGhlbGxvIHdvcmxkCg=="
)

func TestGenerateContainerConf(t *testing.T) {
	gotEnv, gotEntry, gotCmd := GenerateContainerConf([]string{"echo hello world"}, "windows")
	assert.Equal(t, windowsScriptBase64, gotEnv["CI_SCRIPT"])
	assert.Equal(t, "c:\\root", gotEnv["HOME"])
	assert.Equal(t, "powershell.exe", gotEnv["SHELL"])
	assert.Equal(t, []string{"powershell.exe", "-noprofile", "-noninteractive", "-command"}, gotEntry)
	assert.Equal(t, []string{"[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String($Env:CI_SCRIPT)) | iex"}, gotCmd)
	gotEnv, gotEntry, gotCmd = GenerateContainerConf([]string{"echo hello world"}, "linux")
	assert.Equal(t, posixScriptBase64, gotEnv["CI_SCRIPT"])
	assert.Equal(t, "/root", gotEnv["HOME"])
	assert.Equal(t, "/bin/sh", gotEnv["SHELL"])
	assert.Equal(t, []string{"/bin/sh", "-c"}, gotEntry)
	assert.Equal(t, []string{"echo $CI_SCRIPT | base64 -d | /bin/sh -e"}, gotCmd)
}

func TestGenerateNetRCCommand(t *testing.T) {
	posix := GenerateNetRCCommand(DefaultPosixEntry)
	assert.EqualValues(t, []string{"echo CmlmIFsgLW4gIiRDSV9ORVRSQ19NQUNISU5FIiBdOyB0aGVuCmNhdCA8PEVPRiA+ICRIT01FLy5uZXRyYwptYWNoaW5lICRDSV9ORVRSQ19NQUNISU5FCmxvZ2luICRDSV9ORVRSQ19VU0VSTkFNRQpwYXNzd29yZCAkQ0lfTkVUUkNfUEFTU1dPUkQKRU9GCmNobW9kIDA2MDAgJEhPTUUvLm5ldHJjCmZpCnVuc2V0IENJX05FVFJDX1VTRVJOQU1FCnVuc2V0IENJX05FVFJDX1BBU1NXT1JECnVuc2V0IENJX1NDUklQVAo= | base64 -d | /bin/sh -e"}, posix)
	windows := GenerateNetRCCommand(DefaultWindowsEntry)
	assert.EqualValues(t, []string{"[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String(\"JEVycm9yQWN0aW9uUHJlZmVyZW5jZSA9ICdTdG9wJzsKJmNtZCAvYyAibWtkaXIgYzpccm9vdCI7CmlmICgkRW52OkNJX05FVFJDX01BQ0hJTkUpIHsKJG5ldHJjPVtzdHJpbmddOjpGb3JtYXQoInswfVxfbmV0cmMiLCRFbnY6SE9NRSk7CiJtYWNoaW5lICRFbnY6Q0lfTkVUUkNfTUFDSElORSIgPj4gJG5ldHJjOwoibG9naW4gJEVudjpDSV9ORVRSQ19VU0VSTkFNRSIgPj4gJG5ldHJjOwoicGFzc3dvcmQgJEVudjpDSV9ORVRSQ19QQVNTV09SRCIgPj4gJG5ldHJjOwp9OwpbRW52aXJvbm1lbnRdOjpTZXRFbnZpcm9ubWVudFZhcmlhYmxlKCJDSV9ORVRSQ19QQVNTV09SRCIsJG51bGwpOwpbRW52aXJvbm1lbnRdOjpTZXRFbnZpcm9ubWVudFZhcmlhYmxlKCJDSV9TQ1JJUFQiLCRudWxsKTs=\")) | iex"}, windows)
	busybox := GenerateNetRCCommand([]string{"/bin/busybox", "-c"})
	assert.Len(t, busybox, 0)
}
