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
	"encoding/base32"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/prometheus/client_golang/prometheus"
	prometheus_auto "github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/cache"
	"go.woodpecker-ci.org/woodpecker/v2/server/queue"
	logService "go.woodpecker-ci.org/woodpecker/v2/server/services/log"
	"go.woodpecker-ci.org/woodpecker/v2/server/services/log/file"
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

func setupMetrics(g *errgroup.Group, _store store.Store) {
	pendingSteps := prometheus_auto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "pending_steps",
		Help:      "Total number of pending pipeline steps.",
	})
	waitingSteps := prometheus_auto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "waiting_steps",
		Help:      "Total number of pipeline waiting on deps.",
	})
	runningSteps := prometheus_auto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "running_steps",
		Help:      "Total number of running pipeline steps.",
	})
	workers := prometheus_auto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "worker_count",
		Help:      "Total number of workers.",
	})
	pipelines := prometheus_auto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "pipeline_total_count",
		Help:      "Total number of pipelines.",
	})
	users := prometheus_auto.NewGauge(prometheus.GaugeOpts{
		Namespace: "woodpecker",
		Name:      "user_count",
		Help:      "Total number of users.",
	})
	repos := prometheus_auto.NewGauge(prometheus.GaugeOpts{
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

func setupLogStore(c *cli.Context, s store.Store) (logService.Service, error) {
	switch c.String("log-store") {
	case "file":
		return file.NewLogStore(c.String("log-store-file-path"))
	default:
		return s, nil
	}
}

const jwtSecretID = "jwt-secret"

func setupJWTSecret(_store store.Store) (string, error) {
	jwtSecret, err := _store.ServerConfigGet(jwtSecretID)
	if errors.Is(err, types.RecordNotExist) {
		jwtSecret := base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		)
		err = _store.ServerConfigSet(jwtSecretID, jwtSecret)
		if err != nil {
			return "", err
		}
		log.Debug().Msg("created jwt secret")
		return jwtSecret, nil
	}

	if err != nil {
		return "", err
	}

	return jwtSecret, nil
}
