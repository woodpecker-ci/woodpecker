package server

import (
	"strings"
	"time"

	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/remote"
)

type configFetcher struct {
	remote_ remote.Remote
	user    *model.User
	repo    *model.Repo
	build   *model.Build
}

func NewConfigFetcher(remote remote.Remote, user *model.User, repo *model.Repo, build *model.Build) *configFetcher {
	return &configFetcher{
		remote_: remote,
		user:    user,
		repo:    repo,
		build:   build,
	}
}

func (cf *configFetcher) Fetch() (files []*remote.FileMeta, err error) {
	var file []byte

	for i := 0; i < 5; i++ {
		select {
		case <-time.After(time.Second * time.Duration(i)):

			// either a file
			if !strings.HasSuffix(cf.repo.Config, "/") {
				file, err = cf.remote_.File(cf.user, cf.repo, cf.build, cf.repo.Config)
				if err == nil {
					return []*remote.FileMeta{{
						Name: cf.repo.Config,
						Data: file,
					}}, nil
				}
			}

			// or a folder
			if strings.HasSuffix(cf.repo.Config, "/") {
				files, err = cf.remote_.Dir(cf.user, cf.repo, cf.build, strings.TrimSuffix(cf.repo.Config, "/"))
				if err == nil {
					return filterPipelineFiles(files), nil
				}
			}

			// or fallback
			if cf.repo.Fallback {
				file, err = cf.remote_.File(cf.user, cf.repo, cf.build, ".drone.yml")
				if err == nil {
					return []*remote.FileMeta{{
						Name: ".drone.yml",
						Data: file,
					}}, nil
				}
			}

			return nil, err
		}
	}

	return []*remote.FileMeta{}, nil
}

func filterPipelineFiles(files []*remote.FileMeta) []*remote.FileMeta {
	var res []*remote.FileMeta

	for _, file := range files {
		if strings.HasSuffix(file.Name, ".yml") || strings.HasSuffix(file.Name, ".yaml") {
			res = append(res, file)
		}
	}

	return res
}
