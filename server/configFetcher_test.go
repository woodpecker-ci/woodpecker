package server

import (
	"testing"

	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/remote/github"
)

func TestFetchGithub(t *testing.T) {
	t.Parallel()

	github, err := github.New(github.Opts{URL: "https://github.com"})
	if err != nil {
		t.Fatal(err)
	}
	configFetcher := &configFetcher{
		remote_: github,
		user:    &model.User{Token: "xxx"},
		repo:    &model.Repo{Owner: "laszlocph", Name: "drone-multipipeline", Config: ".drone"},
		build:   &model.Build{Commit: "89ab7b2d6bfb347144ac7c557e638ab402848fee"},
	}
	configFetcher.Fetch()
}
