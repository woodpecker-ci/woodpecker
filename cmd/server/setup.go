// Copyright 2018 Drone.IO Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/environments"
	"github.com/woodpecker-ci/woodpecker/server/plugins/registry"
	"github.com/woodpecker-ci/woodpecker/server/plugins/secrets"
	"github.com/woodpecker-ci/woodpecker/server/queue"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/remote/bitbucket"
	"github.com/woodpecker-ci/woodpecker/server/remote/bitbucketserver"
	"github.com/woodpecker-ci/woodpecker/server/remote/coding"
	"github.com/woodpecker-ci/woodpecker/server/remote/gitea"
	"github.com/woodpecker-ci/woodpecker/server/remote/github"
	"github.com/woodpecker-ci/woodpecker/server/remote/gitlab"
	"github.com/woodpecker-ci/woodpecker/server/remote/gogs"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/store/datastore"
	"github.com/woodpecker-ci/woodpecker/server/web"
)

func setupStore(c *cli.Context) (store.Store, error) {
	datasource := c.String("datasource")
	driver := c.String("driver")

	if strings.ToLower(driver) == "sqlite3" {
		if new, err := fallbackSqlite3File(datasource); err != nil {
			log.Fatal().Err(err).Msg("fallback to old sqlite3 file failed")
		} else {
			datasource = new
		}
	}

	opts := &datastore.Opts{
		Driver: driver,
		Config: datasource,
	}
	log.Trace().Msgf("setup datastore: %#v", opts)
	return datastore.New(opts)
}

// TODO: Remove this once we are sure users aren't attempting to migrate from Drone to Woodpecker (possibly never)
func fallbackSqlite3File(path string) (string, error) {
	const dockerDefaultPath = "/var/lib/woodpecker/woodpecker.sqlite"
	const dockerDefaultDir = "/var/lib/woodpecker/drone.sqlite"
	const dockerOldPath = "/var/lib/drone/drone.sqlite"
	const standaloneDefault = "woodpecker.sqlite"
	const standaloneOld = "drone.sqlite"

	// custom location was set, use that one
	if path != dockerDefaultPath && path != standaloneDefault {
		return path, nil
	}

	// woodpecker run in standalone mode, file is in same folder but not renamed
	_, err := os.Stat(standaloneOld)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err == nil {
		// rename in same folder should be fine as it should be same docker volume
		log.Warn().Msgf("found sqlite3 file at '%s', rename to '%s'", standaloneOld, standaloneDefault)
		return standaloneDefault, os.Rename(standaloneOld, standaloneDefault)
	}

	// file is at new default("/var/lib/woodpecker/woodpecker.sqlite")
	_, err = os.Stat(dockerDefaultPath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err == nil {
		return path, nil
	}

	// file is in new folder but not renamed
	_, err = os.Stat(dockerDefaultDir)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err == nil {
		// rename in same folder should be fine as it should be same docker volume
		log.Warn().Msgf("found sqlite3 file at '%s', rename to '%s'", dockerDefaultDir, dockerDefaultPath)
		return dockerDefaultPath, os.Rename(dockerDefaultDir, dockerDefaultPath)
	}

	// file is still at old location
	_, err = os.Stat(dockerOldPath)
	if err == nil {
		log.Error().Msgf("found sqlite3 file at deprecated path '%s', please migrate to '%s'", dockerOldPath, dockerDefaultPath)
		return dockerOldPath, nil
	}

	// file does not exist at all
	log.Warn().Msgf("no sqlite3 file found, will create one at '%s'", path)
	return path, nil
}

func setupQueue(c *cli.Context, s store.Store) queue.Queue {
	return model.WithTaskStore(queue.New(), s)
}

func setupSecretService(c *cli.Context, s store.Store) model.SecretService {
	return secrets.New(s)
}

func setupRegistryService(c *cli.Context, s store.Store) model.RegistryService {
	if c.String("docker-config") != "" {
		return registry.Combined(
			registry.New(s),
			registry.Filesystem(c.String("docker-config")),
		)
	} else {
		return registry.New(s)
	}
}

func setupEnvironService(c *cli.Context, s store.Store) model.EnvironService {
	return environments.Filesystem(c.StringSlice("environment"))
}

// SetupRemote helper function to setup the remote from the CLI arguments.
func SetupRemote(c *cli.Context) (remote.Remote, error) {
	switch {
	case c.Bool("github"):
		return setupGithub(c)
	case c.Bool("gitlab"):
		return setupGitlab(c)
	case c.Bool("bitbucket"):
		return setupBitbucket(c)
	case c.Bool("stash"):
		return setupStash(c)
	case c.Bool("gogs"):
		return setupGogs(c)
	case c.Bool("gitea"):
		return setupGitea(c)
	case c.Bool("coding"):
		return setupCoding(c)
	default:
		return nil, fmt.Errorf("version control system not configured")
	}
}

// helper function to setup the Bitbucket remote from the CLI arguments.
func setupBitbucket(c *cli.Context) (remote.Remote, error) {
	opts := &bitbucket.Opts{
		Client: c.String("bitbucket-client"),
		Secret: c.String("bitbucket-secret"),
	}
	log.Trace().Msgf("Remote (bitbucket) opts: %#v", opts)
	return bitbucket.New(opts)
}

// helper function to setup the Gogs remote from the CLI arguments.
func setupGogs(c *cli.Context) (remote.Remote, error) {
	opts := gogs.Opts{
		URL:         c.String("gogs-server"),
		Username:    c.String("gogs-git-username"),
		Password:    c.String("gogs-git-password"),
		PrivateMode: c.Bool("gogs-private-mode"),
		SkipVerify:  c.Bool("gogs-skip-verify"),
	}
	log.Trace().Msgf("Remote (gogs) opts: %#v", opts)
	return gogs.New(opts)
}

// helper function to setup the Gitea remote from the CLI arguments.
func setupGitea(c *cli.Context) (remote.Remote, error) {
	opts := gitea.Opts{
		URL:         c.String("gitea-server"),
		Context:     c.String("gitea-context"),
		Username:    c.String("gitea-git-username"),
		Password:    c.String("gitea-git-password"),
		Client:      c.String("gitea-client"),
		Secret:      c.String("gitea-secret"),
		PrivateMode: c.Bool("gitea-private-mode"),
		SkipVerify:  c.Bool("gitea-skip-verify"),
	}
	if len(opts.URL) == 0 {
		log.Fatal().Msg("WOODPECKER_GITEA_URL must be set")
	}
	log.Trace().Msgf("Remote (gitea) opts: %#v", opts)
	return gitea.New(opts)
}

// helper function to setup the Stash remote from the CLI arguments.
func setupStash(c *cli.Context) (remote.Remote, error) {
	opts := bitbucketserver.Opts{
		URL:               c.String("stash-server"),
		Username:          c.String("stash-git-username"),
		Password:          c.String("stash-git-password"),
		ConsumerKey:       c.String("stash-consumer-key"),
		ConsumerRSA:       c.String("stash-consumer-rsa"),
		ConsumerRSAString: c.String("stash-consumer-rsa-string"),
		SkipVerify:        c.Bool("stash-skip-verify"),
	}
	log.Trace().Msgf("Remote (bitbucketserver) opts: %#v", opts)
	return bitbucketserver.New(opts)
}

// helper function to setup the Gitlab remote from the CLI arguments.
func setupGitlab(c *cli.Context) (remote.Remote, error) {
	return gitlab.New(gitlab.Opts{
		URL:          c.String("gitlab-server"),
		ClientID:     c.String("gitlab-client"),
		ClientSecret: c.String("gitlab-secret"),
		Username:     c.String("gitlab-git-username"),
		Password:     c.String("gitlab-git-password"),
		PrivateMode:  c.Bool("gitlab-private-mode"),
		SkipVerify:   c.Bool("gitlab-skip-verify"),
	})
}

// helper function to setup the GitHub remote from the CLI arguments.
func setupGithub(c *cli.Context) (remote.Remote, error) {
	opts := github.Opts{
		URL:         c.String("github-server"),
		Context:     c.String("github-context"),
		Client:      c.String("github-client"),
		Secret:      c.String("github-secret"),
		Scopes:      c.StringSlice("github-scope"),
		Username:    c.String("github-git-username"),
		Password:    c.String("github-git-password"),
		PrivateMode: c.Bool("github-private-mode"),
		SkipVerify:  c.Bool("github-skip-verify"),
		MergeRef:    c.Bool("github-merge-ref"),
	}
	log.Trace().Msgf("Remote (github) opts: %#v", opts)
	return github.New(opts)
}

// helper function to setup the Coding remote from the CLI arguments.
func setupCoding(c *cli.Context) (remote.Remote, error) {
	opts := coding.Opts{
		URL:        c.String("coding-server"),
		Client:     c.String("coding-client"),
		Secret:     c.String("coding-secret"),
		Scopes:     c.StringSlice("coding-scope"),
		Machine:    c.String("coding-git-machine"),
		Username:   c.String("coding-git-username"),
		Password:   c.String("coding-git-password"),
		SkipVerify: c.Bool("coding-skip-verify"),
	}
	log.Trace().Msgf("Remote (coding) opts: %#v", opts)
	return coding.New(opts)
}

func setupTree(c *cli.Context) *gin.Engine {
	tree := gin.New()
	web.New(
		web.WithSync(time.Hour*72),
		web.WithDocs(c.String("docs")),
	).Register(tree)
	return tree
}

func before(c *cli.Context) error { return nil }

func setupMetrics(g *errgroup.Group, store_ store.Store) {
	pendingJobs := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "pending_jobs",
		Help:      "Total number of pending build processes.",
	})
	waitingJobs := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "waiting_jobs",
		Help:      "Total number of builds waiting on deps.",
	})
	runningJobs := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "running_jobs",
		Help:      "Total number of running build processes.",
	})
	workers := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "worker_count",
		Help:      "Total number of workers.",
	})
	builds := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "build_total_count",
		Help:      "Total number of builds.",
	})
	users := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "user_count",
		Help:      "Total number of users.",
	})
	repos := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "repo_count",
		Help:      "Total number of repos.",
	})

	g.Go(func() error {
		for {
			stats := server.Config.Services.Queue.Info(nil)
			pendingJobs.Set(float64(stats.Stats.Pending))
			waitingJobs.Set(float64(stats.Stats.WaitingOnDeps))
			runningJobs.Set(float64(stats.Stats.Running))
			workers.Set(float64(stats.Stats.Workers))
			time.Sleep(500 * time.Millisecond)
		}
	})
	g.Go(func() error {
		for {
			repoCount, _ := store_.GetRepoCount()
			userCount, _ := store_.GetUserCount()
			buildCount, _ := store_.GetBuildCount()
			builds.Set(float64(buildCount))
			users.Set(float64(userCount))
			repos.Set(float64(repoCount))
			time.Sleep(10 * time.Second)
		}
	})
}
