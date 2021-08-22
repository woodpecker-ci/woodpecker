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
			dir, direrr := cf.remote_.Dir(cf.user, cf.repo, cf.build, strings.TrimSuffix(cf.repo.Config, "/"))

			if direrr == nil {
				return dir, nil
			} else if !cf.repo.Fallback {
				return nil, direrr
			}

			// or fallback
			file, fileerr = cf.remote_.File(cf.user, cf.repo, cf.build, ".drone.yml")
			if fileerr != nil {
				return nil, fileerr
			}

			return []*remote.FileMeta{{
				Name: cf.repo.Config,
				Data: file,
			}}, nil
		}
	}
	return []*remote.FileMeta{}, nil
}
