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
	"crypto"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/cache"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/environments"
	"github.com/woodpecker-ci/woodpecker/server/plugins/registry"
	"github.com/woodpecker-ci/woodpecker/server/plugins/secrets"
	"github.com/woodpecker-ci/woodpecker/server/queue"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/store/datastore"
	"github.com/woodpecker-ci/woodpecker/server/store/types"
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
			log.Debug().Msgf("server has sqlite3 support")
		} else {
			log.Debug().Msgf("server was built without sqlite3 support!")
		}
	}

	if !datastore.SupportedDriver(driver) {
		log.Fatal().Msgf("database driver '%s' not supported", driver)
	}

	if driver == "sqlite3" {
		if err := checkSqliteFileExist(datasource); err != nil {
			log.Fatal().Err(err).Msg("check sqlite file")
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
		log.Fatal().Err(err).Msg("could not open datastore")
	}

	if err := store.Migrate(); err != nil {
		log.Fatal().Err(err).Msg("could not migrate datastore")
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

func setupSecretService(c *cli.Context, s model.SecretStore) model.SecretService {
	return secrets.New(c.Context, s)
}

func setupRegistryService(c *cli.Context, s store.Store) model.RegistryService {
	if c.String("docker-config") != "" {
		return registry.Combined(
			registry.New(s),
			registry.Filesystem(c.String("docker-config")),
		)
	}
	return registry.New(s)
}

func setupEnvironService(c *cli.Context, _ store.Store) model.EnvironService {
	return environments.Parse(c.StringSlice("environment"))
}

func setupMembershipService(_ *cli.Context, _store store.Store) cache.MembershipService {
	return cache.NewMembershipService(_store)
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

// setupSignatureKeys generate or load key pair to sign webhooks requests (i.e. used for extensions)
func setupSignatureKeys(_store store.Store) (crypto.PrivateKey, crypto.PublicKey) {
	privKeyID := "signature-private-key"

	privKey, err := _store.ServerConfigGet(privKeyID)
	if errors.Is(err, types.RecordNotExist) {
		_, privKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to generate private key")
			return nil, nil
		}
		err = _store.ServerConfigSet(privKeyID, hex.EncodeToString(privKey))
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to generate private key")
			return nil, nil
		}
		log.Debug().Msg("Created private key")
		return privKey, privKey.Public()
	} else if err != nil {
		log.Fatal().Err(err).Msgf("Failed to load private key")
		return nil, nil
	} else {
		privKeyStr, err := hex.DecodeString(privKey)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to decode private key")
			return nil, nil
		}
		privKey := ed25519.PrivateKey(privKeyStr)
		return privKey, privKey.Public()
	}
}
