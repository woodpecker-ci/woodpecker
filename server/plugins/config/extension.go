package config

import (
	"context"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
)

type Extension interface {
	IsConfigured() bool
	FetchConfig(ctx context.Context, repo *model.Repo, pipeline *model.Pipeline, currentFileMeta []*remote.FileMeta) (configData []*remote.FileMeta, useOld bool, err error)
}
