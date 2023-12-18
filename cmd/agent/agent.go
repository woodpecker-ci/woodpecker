// Copyright 2023 Woodpecker Authors
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
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/tevino/abool/v2"
	"github.com/urfave/cli/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpccredentials "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"go.woodpecker-ci.org/woodpecker/v2/agent"
	agentRpc "go.woodpecker-ci.org/woodpecker/v2/agent/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/cmd/common"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/shared/addon"
	addonTypes "go.woodpecker-ci.org/woodpecker/v2/shared/addon/types"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

func run(c *cli.Context) error {
	common.SetupGlobalLogger(c, true)

	agentConfigPath := c.String("agent-config")
	hostname := c.String("hostname")
	if len(hostname) == 0 {
		hostname, _ = os.Hostname()
	}

	counter.Polling = c.Int("max-workflows")
	counter.Running = 0

	if c.Bool("healthcheck") {
		go func() {
			if err := http.ListenAndServe(c.String("healthcheck-addr"), nil); err != nil {
				log.Error().Msgf("cannot listen on address %s: %v", c.String("healthcheck-addr"), err)
			}
		}()
	}

	var transport grpc.DialOption
	if c.Bool("grpc-secure") {
		transport = grpc.WithTransportCredentials(grpccredentials.NewTLS(&tls.Config{InsecureSkipVerify: c.Bool("grpc-skip-insecure")}))
	} else {
		transport = grpc.WithTransportCredentials(insecure.NewCredentials())
	}

	authConn, err := grpc.Dial(
		c.String("server"),
		transport,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    c.Duration("grpc-keepalive-time"),
			Timeout: c.Duration("grpc-keepalive-timeout"),
		}),
	)
	if err != nil {
		return err
	}
	defer authConn.Close()

	agentConfig := readAgentConfig(agentConfigPath)

	agentToken := c.String("grpc-token")
	authClient := agentRpc.NewAuthGrpcClient(authConn, agentToken, agentConfig.AgentID)
	authInterceptor, err := agentRpc.NewAuthInterceptor(authClient, 30*time.Minute) //nolint: gomnd
	if err != nil {
		return err
	}

	conn, err := grpc.Dial(
		c.String("server"),
		transport,
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    c.Duration("grpc-keepalive-time"),
			Timeout: c.Duration("grpc-keepalive-timeout"),
		}),
		grpc.WithUnaryInterceptor(authInterceptor.Unary()),
		grpc.WithStreamInterceptor(authInterceptor.Stream()),
	)
	if err != nil {
		return err
	}
	defer conn.Close()

	client := agentRpc.NewGrpcClient(conn)

	sigterm := abool.New()
	ctx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("hostname", hostname),
	)

	agentConfigPersisted := abool.New()
	ctx = utils.WithContextSigtermCallback(ctx, func() {
		log.Info().Msg("Termination signal is received, shutting down")
		sigterm.Set()

		// Remove stateless agents from server
		if agentConfigPersisted.IsNotSet() {
			log.Debug().Msg("Unregistering agent from server")
			err := client.UnregisterAgent(ctx)
			if err != nil {
				log.Err(err).Msg("Failed to unregister agent from server")
			}
		}
	})

	// check if grpc server version is compatible with agent
	grpcServerVersion, err := client.Version(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not get grpc server version")
		return err
	}
	if grpcServerVersion.GrpcVersion != agentRpc.ClientGrpcVersion {
		err := errors.New("GRPC version mismatch")
		log.Error().Err(err).Msgf("Server version %s does report grpc version %d but we only understand %d",
			grpcServerVersion.ServerVersion,
			grpcServerVersion.GrpcVersion,
			agentRpc.ClientGrpcVersion)
		return err
	}

	var wg sync.WaitGroup
	parallel := c.Int("max-workflows")
	wg.Add(parallel)

	// new engine
	backendCtx := context.WithValue(ctx, types.CliContext, c)
	backendEngine, err := getBackendEngine(backendCtx, c.String("backend-engine"), c.StringSlice("addons"))
	if err != nil {
		return err
	}

	if !backendEngine.IsAvailable(backendCtx) {
		log.Error().Str("engine", backendEngine.Name()).Msg("selected backend engine is unavailable")
		return fmt.Errorf("selected backend engine %s is unavailable", backendEngine.Name())
	}

	// load engine (e.g. init api client)
	engInfo, err := backendEngine.Load(backendCtx)
	if err != nil {
		log.Error().Err(err).Msg("cannot load backend engine")
		return err
	}
	log.Debug().Msgf("loaded %s backend engine", backendEngine.Name())

	agentConfig.AgentID, err = client.RegisterAgent(ctx, engInfo.Platform, backendEngine.Name(), version.String(), parallel)
	if err != nil {
		return err
	}

	if agentConfigPath != "" {
		if err := writeAgentConfig(agentConfig, agentConfigPath); err == nil {
			agentConfigPersisted.Set()
		}
	}

	labels := map[string]string{
		"hostname": hostname,
		"platform": engInfo.Platform,
		"backend":  backendEngine.Name(),
		"repo":     "*", // allow all repos by default
	}

	if err := stringSliceAddToMap(c.StringSlice("filter"), labels); err != nil {
		return err
	}

	filter := rpc.Filter{
		Labels: labels,
	}

	log.Debug().Msgf("Agent registered with ID %d", agentConfig.AgentID)

	go func() {
		for {
			if sigterm.IsSet() {
				log.Debug().Msg("Terminating health reporting")
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
		i := i
		go func() {
			defer wg.Done()

			r := agent.NewRunner(client, filter, hostname, counter, &backendEngine)
			log.Debug().Msgf("created new runner %d", i)

			for {
				if sigterm.IsSet() {
					log.Debug().Msgf("terminating runner %d", i)
					return
				}

				log.Debug().Msg("polling new steps")
				if err := r.Run(ctx); err != nil {
					log.Error().Err(err).Msg("pipeline done with error")
					return
				}
			}
		}()
	}

	log.Info().Msgf(
		"Starting Woodpecker agent with version '%s' and backend '%s' using platform '%s' running up to %d pipelines in parallel",
		version.String(), backendEngine.Name(), engInfo.Platform, parallel)

	wg.Wait()
	return nil
}

func getBackendEngine(backendCtx context.Context, backendName string, addons []string) (types.Backend, error) {
	addonBackend, err := addon.Load[types.Backend](addons, addonTypes.TypeBackend)
	if err != nil {
		log.Error().Err(err).Msg("cannot load backend addon")
		return nil, err
	}
	if addonBackend != nil {
		return addonBackend.Value, nil
	}

	engine, err := backend.FindBackend(backendCtx, backendName)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find backend engine '%s'", backendName)
		return nil, err
	}
	return engine, nil
}

func runWithRetry(context *cli.Context) error {
	retryCount := context.Int("connect-retry-count")
	retryDelay := context.Duration("connect-retry-delay")
	var err error
	for i := 0; i < retryCount; i++ {
		if err = run(context); status.Code(err) == codes.Unavailable {
			log.Warn().Err(err).Msg(fmt.Sprintf("cannot connect to server, retrying in %v", retryDelay))
			time.Sleep(retryDelay)
		} else {
			break
		}
	}
	return err
}

func stringSliceAddToMap(sl []string, m map[string]string) error {
	if m == nil {
		m = make(map[string]string)
	}
	//nolint: gomnd
	for _, v := range sl {
		parts := strings.SplitN(v, "=", 2)
		switch len(parts) {
		case 2:
			m[parts[0]] = parts[1]
		case 1:
			return fmt.Errorf("key '%s' does not have a value assigned", parts[0])
		default:
			return fmt.Errorf("empty string in slice")
		}
	}
	return nil
}
