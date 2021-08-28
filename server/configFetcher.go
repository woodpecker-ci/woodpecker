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

func (cf *configFetcher) Fetch() ([]*remote.FileMeta, error) {
	for i := 0; i < 5; i++ {
		select {
		case <-time.After(time.Second * time.Duration(i)):
			// either a file
			file, fileerr := cf.remote_.File(cf.user, cf.repo, cf.build, cf.repo.Config)
			if fileerr == nil {
				return []*remote.FileMeta{{
					Name: cf.repo.Config,
					Data: file,
				}}, nil
			}

			// or a folder
			dir, fileerr := cf.remote_.Dir(cf.user, cf.repo, cf.build, strings.TrimSuffix(cf.repo.Config, "/"))
			if fileerr == nil {
				return filterPipelineFiles(dir), nil
			} else

			// or fallback
			if cf.repo.Fallback {
				file, fileerr = cf.remote_.File(cf.user, cf.repo, cf.build, ".drone.yml")
				if fileerr == nil {
					return []*remote.FileMeta{{
						Name: cf.repo.Config,
						Data: file,
					}}, nil
				}
			}

			return nil, fileerr
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
