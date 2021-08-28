package server_test

import (
	"testing"

	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/remote"
	"github.com/woodpecker-ci/woodpecker/remote/github"
	"github.com/woodpecker-ci/woodpecker/remote/mocks"
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

func TestFilterPipelineFiles(t *testing.T) {
	t.Parallel()

	user := &model.User{Token: "xxx"}
	repo := &model.Repo{Owner: "laszlocph", Name: "drone-multipipeline", Config: ".woodpecker/"}
	build := &model.Build{Commit: "89ab7b2d6bfb347144ac7c557e638ab402848fee"}

	r := new(mocks.Remote)
	r.On("File", user, repo, build, ".woodpecker/").Return(nil, nil)
	r.On("Dir", user, repo, build, ".woodpecker/").Return([]*remote.FileMeta{
		{
			Name: ".woodpecker/test.yml",
			Data: []byte{},
		}, {
			Name: ".woodpecker/text.txt",
			Data: []byte{},
		},
		{
			Name: ".woodpecker/image.png",
			Data: []byte{},
		},
	}, nil)

	configFetcher := server.NewConfigFetcher(
		r,
		user,
		repo,
		build,
	)
	files, err := configFetcher.Fetch()
	if err != nil {
		t.Fatal("uff", err)
	}

	if len(files) != 1 {
		t.Fatal()
	}
}
