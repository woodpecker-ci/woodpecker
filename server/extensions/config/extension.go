package config

import (
	"context"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
)

type Extension interface {
	FetchConfig(ctx context.Context, user *model.User, repo *model.Repo, build *model.Build) (configData []*remote.FileMeta, err error)
}
