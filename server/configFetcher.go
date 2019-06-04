package server

import (
	"time"

	"github.com/laszlocph/drone-oss-08/model"
	"github.com/laszlocph/drone-oss-08/remote"
)

type configFetcher struct {
	remote_ remote.Remote
	user    *model.User
	repo    *model.Repo
	build   *model.Build
}

func (cf *configFetcher) Fetch() ([]*remote.FileMeta, error) {
	for i := 0; i < 5; i++ {
		select {
		case <-time.After(time.Second * time.Duration(i)):
			file, err := cf.remote_.File(cf.user, cf.repo, cf.build, cf.repo.Config) // either a file
			if err == nil {
				return []*remote.FileMeta{&remote.FileMeta{
					Name: cf.repo.Config,
					Data: file,
				}}, nil
			}

			dir, err := cf.remote_.Dir(cf.user, cf.repo, cf.build, cf.repo.Config) // or a folder
			if err != nil {
				return nil, err
			}
			return dir, nil
		}
	}
	return []*remote.FileMeta{}, nil
}
