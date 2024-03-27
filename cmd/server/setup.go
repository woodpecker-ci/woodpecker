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
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/cache"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge"
	"go.woodpecker-ci.org/woodpecker/v2/server/forge/loader"
	"go.woodpecker-ci.org/woodpecker/v2/server/model"
	"go.woodpecker-ci.org/woodpecker/v2/server/queue"
	"go.woodpecker-ci.org/woodpecker/v2/server/store"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/datastore"
	"go.woodpecker-ci.org/woodpecker/v2/server/store/types"
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

func setupMembershipService(_ *cli.Context, _store store.Store) cache.MembershipService {
	return cache.NewMembershipService(_store)
}

func setupForgeService(c *cli.Context, _store store.Store) (forge.ForgeService, error) {
	var forgeType string
	additionalOptions := map[string]any{}

	switch {
	case c.Bool("github"):
		forgeType = "github"
		additionalOptions["merge-ref"] = c.Bool("github-merge-ref")
	case c.Bool("gitlab"):
		forgeType = "gitlab"
	case c.Bool("gitea"):
		forgeType = "gitea"
		additionalOptions["oauth-server"] = c.String("gitea-oauth-server")
	case c.Bool("bitbucket"):
		forgeType = "bitbucket"
	case c.Bool("bitbucket-dc"):
		forgeType = "bitbucket-dc"
		additionalOptions["git-username"] = c.String("bitbucket-dc-git-username")
		additionalOptions["git-password"] = c.String("bitbucket-dc-git-password")
	default:
		return nil, errors.New("forge not configured")
	}

	_forge, err := _store.ForgeGet(0)
	if errors.Is(err, types.RecordNotExist) {
		return nil, err
	}
	forgeExists := err == nil
	if _forge == nil {
		_forge = &model.Forge{
			ID: 0,
		}
	}

	_forge.Type = forgeType
	_forge.Client = c.String("forge-oauth-client")
	_forge.ClientSecret = c.String("forge-oauth-secret")
	_forge.URL = c.String("forge-url")
	_forge.SkipVerify = c.Bool("forge-skip-verify")
	_forge.AdditionalOptions = additionalOptions

	if forgeExists {
		err := _store.ForgeUpdate(_forge)
		if err != nil {
			return nil, err
		}
	} else {
		err := _store.ForgeCreate(_forge)
		if err != nil {
			return nil, err
		}
	}

	return loader.NewForgeService(_store), nil
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
