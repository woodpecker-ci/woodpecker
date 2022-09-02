package compiler

import (
	"encoding/base64"
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

echo + "echo \${PATH}"
echo ${PATH}

echo + "go build"
go build

echo + "go test"
go test

`,
		},
	}
	for _, test := range testdata {
		script := generateScriptPosix(test.from)
		decoded, _ := base64.StdEncoding.DecodeString(script)
		got := string(decoded)
		assert.EqualValues(t, got, test.want, "Want encoded script for %s", test.from)
	}
}
