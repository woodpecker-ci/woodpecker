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
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/tevino/abool"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	grpccredentials "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"github.com/woodpecker-ci/woodpecker/agent"
	"github.com/woodpecker-ci/woodpecker/pipeline/backend/docker"
	"github.com/woodpecker-ci/woodpecker/pipeline/rpc"
)

func loop(c *cli.Context) error {
	filter := rpc.Filter{
		Labels: map[string]string{
			"platform": c.String("platform"),
		},
		Expr: c.String("filter"),
	}

	hostname := c.String("hostname")
	if len(hostname) == 0 {
		hostname, _ = os.Hostname()
	}

	if c.Bool("pretty") {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:     os.Stderr,
				NoColor: c.Bool("nocolor"),
			},
		)
	}

	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	if c.Bool("debug") {
		if c.IsSet("debug") {
			log.Warn().Msg("--debug is deprecated, use --log-level instead")
		}
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if c.IsSet("log-level") {
		logLevelFlag := c.String("log-level")
		lvl, err := zerolog.ParseLevel(logLevelFlag)
		if err != nil {
			log.Fatal().Msgf("unknown logging level: %s", logLevelFlag)
		}
		zerolog.SetGlobalLevel(lvl)
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

	var transport = grpc.WithInsecure()

	if c.Bool("secure-grpc") {
		transport = grpc.WithTransportCredentials(grpccredentials.NewTLS(&tls.Config{InsecureSkipVerify: c.Bool("skip-insecure-grpc")}))
	}

	conn, err := grpc.Dial(
		c.String("server"),
		transport,
		grpc.WithPerRPCCredentials(&credentials{
			username: c.String("username"),
			password: c.String("password"),
		}),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    c.Duration("keepalive-time"),
			Timeout: c.Duration("keepalive-timeout"),
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
	ctx = WithContextFunc(ctx, func() {
		println("ctrl+c received, terminating process")
		sigterm.Set()
	})

	var wg sync.WaitGroup
	parallel := c.Int("max-procs")
	wg.Add(parallel)

	for i := 0; i < parallel; i++ {
		go func() {
			defer wg.Done()
			for {
				if sigterm.IsSet() {
					return
				}

				// new docker engine
				engine, err := docker.NewEnv()
				if err != nil {
					log.Error().Err(err).Msg("cannot create docker client")
					return
				}

				r := agent.NewRunner(client, filter, hostname, counter, &engine)
				if err := r.Run(ctx); err != nil {
					log.Error().Err(err).Msg("pipeline done with error")
					return
				}
			}
		}()
	}

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
