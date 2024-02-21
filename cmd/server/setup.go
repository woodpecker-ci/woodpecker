// Copyright 2022 Woodpecker Authors
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
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/cache"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/bitbucket"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/bitbucketdatacenter"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/gitea"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/github"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/gitlab"
	"go.woodpecker-ci.org/woodpecker/v2/server/queue"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/datastore"
	"go.woodpecker-ci.org/woodpecker/v2/shared/addon"
	addonTypes "go.woodpecker-ci.org/woodpecker/v2/shared/addon/types"
)

func setupStore(c *cli.Context) (store.Store, error) {
	datasource := c.String("datasource")
	driver := c.String("driver")
	xorm := store.XORM{
		Log:     c.Bool("log-xorm"),
		ShowSQL: c.Bool("log-xorm-sql"),
	}

	if driver == "sqlite3" {
		if datastore.SupportedDriver("sqlite3") {
			log.Debug().Msg("server has sqlite3 support")
		} else {
			log.Debug().Msg("server was built without sqlite3 support!")
		}
	}

	if !datastore.SupportedDriver(driver) {
		return nil, fmt.Errorf("database driver '%s' not supported", driver)
	}

	if driver == "sqlite3" {
		if err := checkSqliteFileExist(datasource); err != nil {
			return nil, fmt.Errorf("check sqlite file: %w", err)
		}
	}

	opts := &store.Opts{
		Driver: driver,
		Config: datasource,
		XORM:   xorm,
	}
	log.Trace().Msgf("setup datastore: %#v", *opts)
	store, err := datastore.NewEngine(opts)
	if err != nil {
		return nil, fmt.Errorf("could not open datastore: %w", err)
	}

	if err := store.Migrate(c.Bool("migrations-allow-long")); err != nil {
		return nil, fmt.Errorf("could not migrate datastore: %w", err)
	}

	return store, nil
}

func checkSqliteFileExist(path string) error {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		log.Warn().Msgf("no sqlite3 file found, will create one at '%s'", path)
		return nil
	}
	return err
}

func setupQueue(c *cli.Context, s store.Store) queue.Queue {
	return queue.WithTaskStore(queue.New(c.Context), s)
}

func setupMembershipService(_ *cli.Context, r forge.Forge) cache.MembershipService {
	return cache.NewMembershipService(r)
}

// setupForge helper function to set up the forge from the CLI arguments.
func setupForge(c *cli.Context) (forge.Forge, error) {
	addonForge, err := addon.Load[forge.Forge](c.StringSlice("addons"), addonTypes.TypeForge)
	if err != nil {
		return nil, err
	}
	if addonForge != nil {
		return addonForge.Value, nil
	}

	switch {
	case c.Bool("github"):
		return setupGitHub(c)
	case c.Bool("gitlab"):
		return setupGitLab(c)
	case c.Bool("bitbucket"):
		return setupBitbucket(c)
	case c.Bool("bitbucket-dc"):
		return setupBitbucketDatacenter(c)
	case c.Bool("gitea"):
		return setupGitea(c)
	default:
		return nil, fmt.Errorf("version control system not configured")
	}
}

// setupBitbucket helper function to setup the Bitbucket forge from the CLI arguments.
func setupBitbucket(c *cli.Context) (forge.Forge, error) {
	opts := &bitbucket.Opts{
		Client: c.String("bitbucket-client"),
		Secret: c.String("bitbucket-secret"),
	}
	log.Trace().Msgf("forge (bitbucket) opts: %#v", opts)
	return bitbucket.New(opts)
}

// setupGitea helper function to setup the Gitea forge from the CLI arguments.
func setupGitea(c *cli.Context) (forge.Forge, error) {
	server, err := url.Parse(c.String("gitea-server"))
	if err != nil {
		return nil, err
	}
	opts := gitea.Opts{
		URL:        strings.TrimRight(server.String(), "/"),
		Client:     c.String("gitea-client"),
		Secret:     c.String("gitea-secret"),
		SkipVerify: c.Bool("gitea-skip-verify"),
	}
	if len(opts.URL) == 0 {
		return nil, fmt.Errorf("WOODPECKER_GITEA_URL must be set")
	}
	log.Trace().Msgf("forge (gitea) opts: %#v", opts)
	return gitea.New(opts)
}

// setupBitbucketDatacenter helper function to setup the Bitbucket DataCenter/Server forge from the CLI arguments.
func setupBitbucketDatacenter(c *cli.Context) (forge.Forge, error) {
	opts := bitbucketdatacenter.Opts{
		URL:          c.String("bitbucket-dc-server"),
		Username:     c.String("bitbucket-dc-git-username"),
		Password:     c.String("bitbucket-dc-git-password"),
		ClientID:     c.String("bitbucket-dc-client-id"),
		ClientSecret: c.String("bitbucket-dc-client-secret"),
	}
	log.Trace().Msgf("Forge (bitbucketdatacenter) opts: %#v", opts)
	return bitbucketdatacenter.New(opts)
}

// setupGitLab helper function to setup the GitLab forge from the CLI arguments.
func setupGitLab(c *cli.Context) (forge.Forge, error) {
	return gitlab.New(gitlab.Opts{
		URL:          c.String("gitlab-server"),
		ClientID:     c.String("gitlab-client"),
		ClientSecret: c.String("gitlab-secret"),
		SkipVerify:   c.Bool("gitlab-skip-verify"),
	})
}

// setupGitHub helper function to setup the GitHub forge from the CLI arguments.
func setupGitHub(c *cli.Context) (forge.Forge, error) {
	opts := github.Opts{
		URL:        c.String("github-server"),
		Client:     c.String("github-client"),
		Secret:     c.String("github-secret"),
		SkipVerify: c.Bool("github-skip-verify"),
		MergeRef:   c.Bool("github-merge-ref"),
	}
	log.Trace().Msgf("forge (github) opts: %#v", opts)
	return github.New(opts)
}

func setupMetrics(g *errgroup.Group, _store store.Store) {
	pendingSteps := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "pending_steps",
		Help:      "Total number of pending pipeline steps.",
	})
	waitingSteps := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "waiting_steps",
		Help:      "Total number of pipeline waiting on deps.",
	})
	runningSteps := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "running_steps",
		Help:      "Total number of running pipeline steps.",
	})
	workers := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "worker_count",
		Help:      "Total number of workers.",
	})
	pipelines := promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "pipeline_total_count",
		Help:      "Total number of pipelines.",
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
			stats := server.Config.Services.Queue.Info(context.TODO())
			pendingSteps.Set(float64(stats.Stats.Pending))
			waitingSteps.Set(float64(stats.Stats.WaitingOnDeps))
			runningSteps.Set(float64(stats.Stats.Running))
			workers.Set(float64(stats.Stats.Workers))
			time.Sleep(500 * time.Millisecond)
		}
	})
	g.Go(func() error {
		for {
			repoCount, _ := _store.GetRepoCount()
			userCount, _ := _store.GetUserCount()
			pipelineCount, _ := _store.GetPipelineCount()
			pipelines.Set(float64(pipelineCount))
			users.Set(float64(userCount))
			repos.Set(float64(repoCount))
			time.Sleep(10 * time.Second)
		}
	})
}
