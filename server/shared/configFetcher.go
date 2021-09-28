package shared

import (
	"context"
	"errors"
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

// Fetch pipeline config from source forge
func (cf *configFetcher) Fetch(ctx context.Context) (files []*remote.FileMeta, err error) {
	logrus.Tracef("Start Fetching config for '%s'", cf.repo.FullName)

	// try to fetch 3 times, timeout is one second longer each time
	for i := 0; i < 3; i++ {
		files, err = cf.fetch(ctx, time.Second*time.Duration(i), strings.TrimSpace(cf.repo.Config))
		if errors.Is(err, context.DeadlineExceeded) {
			continue
		}
		return
	}
	return
}

// fetch config by timeout
// TODO: dedupe code
func (cf *configFetcher) fetch(c context.Context, timeout time.Duration, config string) ([]*remote.FileMeta, error) {
	ctx, cancel := context.WithTimeout(c, timeout)
	defer cancel()

	if len(config) > 0 {
		logrus.Tracef("ConfigFetch[%s]: use user config '%s'", cf.repo.FullName, config)
		// either a file
		if !strings.HasSuffix(config, "/") {
			file, err := cf.remote_.File(ctx, cf.user, cf.repo, cf.build, config)
			if err == nil && len(file) != 0 {
				logrus.Tracef("ConfigFetch[%s]: found file '%s'", cf.repo.FullName, config)
				return []*remote.FileMeta{{
					Name: config,
					Data: file,
				}}, nil
			}
		}

		// or a folder
		files, err := cf.remote_.Dir(ctx, cf.user, cf.repo, cf.build, strings.TrimSuffix(config, "/"))
		if err == nil && len(files) != 0 {
			logrus.Tracef("ConfigFetch[%s]: found %d files in '%s'", cf.repo.FullName, len(files), config)
			return filterPipelineFiles(files), nil
		}

		return nil, fmt.Errorf("config '%s' not found: %s", config, err)
	}

	logrus.Tracef("ConfigFetch[%s]: user did not defined own config follow default procedure", cf.repo.FullName)
	// no user defined config so try .woodpecker/*.yml -> .woodpecker.yml -> .drone.yml

	// test .woodpecker/ folder
	// if folder is not supported we will get a "Not implemented" error and continue
	config = ".woodpecker"
	files, err := cf.remote_.Dir(ctx, cf.user, cf.repo, cf.build, config)
	files = filterPipelineFiles(files)
	if err == nil && len(files) != 0 {
		logrus.Tracef("ConfigFetch[%s]: found %d files in '%s'", cf.repo.FullName, len(files), config)
		return files, nil
	}

	config = ".woodpecker.yml"
	file, err := cf.remote_.File(ctx, cf.user, cf.repo, cf.build, config)
	if err == nil && len(file) != 0 {
		logrus.Tracef("ConfigFetch[%s]: found file '%s'", cf.repo.FullName, config)
		return []*remote.FileMeta{{
			Name: config,
			Data: file,
		}}, nil
	}

	config = ".drone.yml"
	file, err = cf.remote_.File(ctx, cf.user, cf.repo, cf.build, config)
	if err == nil && len(file) != 0 {
		logrus.Tracef("ConfigFetch[%s]: found file '%s'", cf.repo.FullName, config)
		return []*remote.FileMeta{{
			Name: config,
			Data: file,
		}}, nil
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return []*remote.FileMeta{}, fmt.Errorf("ConfigFetcher: Fallback did not found config: %s", err)
	}
}

func filterPipelineFiles(files []*remote.FileMeta) []*remote.FileMeta {
	var res []*remote.FileMeta

	for _, file := range files {
		if strings.HasSuffix(file.Name, ".yml") {
			res = append(res, file)
		}
	}

	return res
}
