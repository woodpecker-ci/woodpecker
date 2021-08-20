package server_test

import (
	"testing"

	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/remote/github"
	"github.com/woodpecker-ci/woodpecker/server"
)

func TestFetchGithub(t *testing.T) {
	t.Parallel()

	github, err := github.New(github.Opts{URL: "https://github.com"})
	if err != nil {
		t.Fatal(err)
	}
	configFetcher := server.NewConfigFetcher(
		github,
		&model.User{Token: "xxx"},
		&model.Repo{Owner: "laszlocph", Name: "drone-multipipeline", Config: ".drone"},
		&model.Build{Commit: "89ab7b2d6bfb347144ac7c557e638ab402848fee"},
	)
	configFetcher.Fetch()
}
