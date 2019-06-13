package server

import (
	"time"

	"github.com/laszlocph/drone-oss-08/model"
	"github.com/laszlocph/drone-oss-08/remote"

	"sort"
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
			file, fileerr := cf.remote_.File(cf.user, cf.repo, cf.build, cf.repo.Config) // either a file
			if fileerr == nil {
				return []*remote.FileMeta{&remote.FileMeta{
					Name: cf.repo.Config,
					Data: file,
				}}, nil
			}

			dir, direrr := cf.remote_.Dir(cf.user, cf.repo, cf.build, ".drone") // or a folder
			if direrr != nil {
				return nil, direrr
			}

			sort.Sort(byName(dir))
			return dir, nil
		}
	}
	return []*remote.FileMeta{}, nil
}

type byName []*remote.FileMeta

func (a byName) Len() int           { return len(a) }
func (a byName) Less(i, j int) bool { return a[i].Name < a[j].Name }
func (a byName) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
