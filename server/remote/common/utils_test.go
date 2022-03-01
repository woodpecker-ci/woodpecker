package common_test

import (
	"testing"

	"github.com/woodpecker-ci/woodpecker/server/remote/common"
)

func Test_Netrc(t *testing.T) {
	host, err := common.ExtractHostFromCloneURL("https://git.example.com/foo/bar.git")
	if err != nil {
		t.Fatal(err)
	}

	if host != "git.example.com" {
		t.Errorf("Expected host to be git.example.com, got %s", host)
	}
}
