package shared

import (
	"fmt"
	"strings"
	"time"

	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/remote"

	"github.com/sirupsen/logrus"
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

// Fetch
// TODO: dedupe code
func (cf *configFetcher) Fetch() (files []*remote.FileMeta, err error) {
	var file []byte
	config := strings.TrimSpace(cf.repo.Config)

	logrus.Tracef("Start Fetching config for '%s'", cf.repo.FullName)

	for i := 0; i < 5; i++ {
		select {
		case <-time.After(time.Second * time.Duration(i)):
			if len(config) > 0 {
				// either a file
				if !strings.HasSuffix(config, "/") {
					file, err = cf.remote_.File(cf.user, cf.repo, cf.build, config)
					if err == nil && len(file) != 0 {
						logrus.Tracef("ConfigFetch[%s]: found file '%s'", cf.repo.FullName, config)
						return []*remote.FileMeta{{
							Name: config,
							Data: file,
						}}, nil
					}
				}

				// or a folder
				files, err = cf.remote_.Dir(cf.user, cf.repo, cf.build, strings.TrimSuffix(config, "/"))
				if err == nil {
					logrus.Tracef("ConfigFetch[%s]: found %d files in '%s'", cf.repo.FullName, len(files), config)
					return filterPipelineFiles(files), nil
				}
			} else {
				logrus.Tracef("ConfigFetch[%s]: user did not defined own config follow default procedure", cf.repo.FullName)
				// no user defined config so try .woodpecker/*.yml -> .woodpecker.yml -> .drone.yml

				// test .woodpecker/ folder
				// if folder is not supported we will get a "Not implemented" error and continue
				config = ".woodpecker"
				files, err = cf.remote_.Dir(cf.user, cf.repo, cf.build, config)
				logrus.Tracef("ConfigFetch[%s]: found %d files in '%s'", cf.repo.FullName, len(files), config)
				files = filterPipelineFiles(files)
				if err == nil && len(files) != 0 {
					return files, nil
				}

				config = ".woodpecker.yml"
				file, err = cf.remote_.File(cf.user, cf.repo, cf.build, config)
				if err == nil && len(file) != 0 {
					logrus.Tracef("ConfigFetch[%s]: found file '%s'", cf.repo.FullName, config)
					return []*remote.FileMeta{{
						Name: ".woodpecker.yml",
						Data: file,
					}}, nil
				}

				config = ".drone.yml"
				file, err = cf.remote_.File(cf.user, cf.repo, cf.build, config)
				if err == nil && len(file) != 0 {
					logrus.Tracef("ConfigFetch[%s]: found file '%s'", cf.repo.FullName, config)
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
