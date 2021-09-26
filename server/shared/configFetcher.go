package shared

import (
	"fmt"
	"strings"
	"time"

	"github.com/woodpecker-ci/woodpecker/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"
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
	config := strings.TrimSpace(cf.repo.Config)

	for i := 0; i < 5; i++ {
		select {
		case <-time.After(time.Second * time.Duration(i)):
			if len(config) > 0 {
				// either a file
				if !strings.HasSuffix(config, "/") {
					file, err = cf.remote_.File(cf.user, cf.repo, cf.build, config)
					if err == nil && len(file) != 0 {
						return []*remote.FileMeta{{
							Name: config,
							Data: file,
						}}, nil
					}
				}

				// or a folder
				files, err = cf.remote_.Dir(cf.user, cf.repo, cf.build, strings.TrimSuffix(config, "/"))
				if err == nil {
					return filterPipelineFiles(files), nil
				}
			} else {
				// no user defined config so try .woodpecker/*.yml -> .woodpecker.yml -> .drone.yml

				// test .woodpecker/ folder
				// if folder is not supported we will get a "Not implemented" error and continue
				files, err = cf.remote_.Dir(cf.user, cf.repo, cf.build, ".woodpecker")
				files = filterPipelineFiles(files)
				if err == nil && len(files) != 0 {
					return files, nil
				}

				file, err = cf.remote_.File(cf.user, cf.repo, cf.build, ".woodpecker.yml")
				if err == nil && len(file) != 0 {
					return []*remote.FileMeta{{
						Name: ".woodpecker.yml",
						Data: file,
					}}, nil
				}

				file, err = cf.remote_.File(cf.user, cf.repo, cf.build, ".drone.yml")
				if err == nil && len(file) != 0 {
					return []*remote.FileMeta{{
						Name: ".drone.yml",
						Data: file,
					}}, nil
				}

				if err == nil && len(files) == 0 {
					return nil, fmt.Errorf("ConfigFetcher: Fallback did not found config")
				}
			}

			// TODO: retry loop is inactive and could maybe be fixed/deleted
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
