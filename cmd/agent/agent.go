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
	"crypto/tls"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tevino/abool"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	grpccredentials "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"github.com/woodpecker-ci/woodpecker/agent"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/types"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
	"github.com/woodpecker-ci/woodpecker/shared/utils"
	"github.com/woodpecker-ci/woodpecker/version"
)

func loop(c *cli.Context) error {
	hostname := c.String("hostname")
	if len(hostname) == 0 {
		hostname, _ = os.Hostname()
	}

	platform := runtime.GOOS + "/" + runtime.GOARCH

	labels := map[string]string{
		"hostname": hostname,
		"platform": platform,
		"repo":     "*", // allow all repos by default
	}

	for _, v := range c.StringSlice("filter") {
		parts := strings.SplitN(v, "=", 2)
		labels[parts[0]] = parts[1]
	}

	filter := rpc.Filter{
		Labels: labels,
	}

	if c.Bool("pretty") {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:     os.Stderr,
				NoColor: c.Bool("nocolor"),
			},
		)
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if c.IsSet("log-level") {
		logLevelFlag := c.String("log-level")
		lvl, err := zerolog.ParseLevel(logLevelFlag)
		if err != nil {
			log.Fatal().Msgf("unknown logging level: %s", logLevelFlag)
		}
		zerolog.SetGlobalLevel(lvl)
	}
	if zerolog.GlobalLevel() <= zerolog.DebugLevel {
		log.Logger = log.With().Caller().Logger()
	}

	counter.Polling = c.Int("max-procs")
	counter.Running = 0

	if c.Bool("healthcheck") {
		go func() {
			if err := http.ListenAndServe(":3000", nil); err != nil {
				log.Error().Msgf("can not listen on port 3000: %v", err)
			}
		}()
	}

	// TODO pass version information to grpc server
	// TODO authenticate to grpc server

	// grpc.Dial(target, ))
	var transport grpc.DialOption
	if c.Bool("grpc-secure") {
		transport = grpc.WithTransportCredentials(grpccredentials.NewTLS(&tls.Config{InsecureSkipVerify: c.Bool("skip-insecure-grpc")}))
	} else {
		transport = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	conn, err := grpc.Dial(
		c.String("server"),
		transport,
		grpc.WithPerRPCCredentials(&credentials{
			username: c.String("grpc-username"),
			password: c.String("grpc-password"),
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    c.Duration("grpc-keepalive-time"),
			Timeout: c.Duration("grpc-keepalive-timeout"),
		}),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := rpc.NewGrpcClient(conn)

	sigterm := abool.New()
	ctx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("hostname", hostname),
	)
	ctx = utils.WithContextSigtermCallback(ctx, func() {
		println("ctrl+c received, terminating process")
		sigterm.Set()
	})

	backend.Init(context.WithValue(ctx, types.CliContext, c))

	var wg sync.WaitGroup
	parallel := c.Int("max-procs")
	wg.Add(parallel)

	// new engine
	engine, err := backend.FindEngine(c.String("backend-engine"))
	if err != nil {
		log.Error().Err(err).Msgf("cannot find backend engine '%s'", c.String("backend-engine"))
		return err
	}

	agentId := int64(-1) // TODO: get agent id from some storage like file etc or use -1 if none exists
	agentId, err = client.RegisterAgent(ctx, agentId, platform, engine.Name(), parallel)
	if err != nil {
		return err
	}
	// TODO: save returned agentId again

	go func() {
		for {
			if sigterm.IsSet() {
				return
			}

			err := client.ReportHealth(ctx)
			if err != nil {
				log.Err(err).Msgf("Failed to report health")
				return
			}

			<-time.After(time.Second * 10)
		}
	}()

	for i := 0; i < parallel; i++ {
		go func() {
			defer wg.Done()

			// load engine (e.g. init api client)
			err = engine.Load()
			if err != nil {
				log.Error().Err(err).Msg("cannot load backend engine")
				return
			}

			r := agent.NewRunner(client, filter, hostname, counter, &engine)

			log.Debug().Msgf("loaded %s backend engine", engine.Name())

			for {
				if sigterm.IsSet() {
					return
				}

				log.Debug().Msg("polling new jobs")
				if err := r.Run(ctx); err != nil {
					log.Error().Err(err).Msg("pipeline done with error")
					return
				}
			}
		}()
	}

	log.Info().Msgf(
		"Starting Woodpecker agent with version '%s' and backend '%s' running up to %d pipelines in parallel",
		version.String(), engine.Name(), parallel)

	wg.Wait()
	return nil
}

type credentials struct {
	username string
	password string
}

func (c *credentials) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"username": c.username,
		"password": c.password,
	}, nil
}

func (c *credentials) RequireTransportSecurity() bool {
	return false
}
