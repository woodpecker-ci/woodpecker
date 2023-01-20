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
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"github.com/woodpecker-ci/woodpecker/pipeline/rpc/proto"
	"github.com/woodpecker-ci/woodpecker/server"
	"github.com/woodpecker-ci/woodpecker/server/cron"
	"github.com/woodpecker-ci/woodpecker/server/forge"
	woodpeckerGrpcServer "github.com/woodpecker-ci/woodpecker/server/grpc"
	"github.com/woodpecker-ci/woodpecker/server/logging"
	"github.com/woodpecker-ci/woodpecker/server/model"
	"github.com/woodpecker-ci/woodpecker/server/plugins/config"
	"github.com/woodpecker-ci/woodpecker/server/pubsub"
	"github.com/woodpecker-ci/woodpecker/server/router"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/web"
	"github.com/woodpecker-ci/woodpecker/version"
	// "github.com/woodpecker-ci/woodpecker/server/plugins/encryption"
	// encryptedStore "github.com/woodpecker-ci/woodpecker/server/plugins/encryption/wrapper/store"
)

func run(c *cli.Context) error {
	if c.Bool("pretty") {
		log.Logger = log.Output(
			zerolog.ConsoleWriter{
				Out:     os.Stderr,
				NoColor: c.Bool("nocolor"),
			},
		)
	}

	// TODO: format output & options to switch to json aka. option to add channels to send logs to
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
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	log.Log().Msgf("LogLevel = %s", zerolog.GlobalLevel().String())

	if c.String("server-host") == "" {
		log.Fatal().Msg("WOODPECKER_HOST is not properly configured")
	}

	if !strings.Contains(c.String("server-host"), "://") {
		log.Fatal().Msg(
			"WOODPECKER_HOST must be <scheme>://<hostname> format",
		)
	}

	if strings.Contains(c.String("server-host"), "://localhost") {
		log.Warn().Msg(
			"WOODPECKER_HOST should probably be publicly accessible (not localhost)",
		)
	}

	if strings.HasSuffix(c.String("server-host"), "/") {
		log.Fatal().Msg(
			"WOODPECKER_HOST must not have trailing slash",
		)
	}

	_forge, err := setupForge(c)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	_store, err := setupStore(c)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	defer func() {
		if err := _store.Close(); err != nil {
			log.Error().Err(err).Msg("could not close store")
		}
	}()

	setupEvilGlobals(c, _store, _forge)

	var g errgroup.Group

	setupMetrics(&g, _store)

	g.Go(func() error {
		return cron.Start(c.Context, _store, _forge)
	})

	// start the grpc server
	g.Go(func() error {
		lis, err := net.Listen("tcp", c.String("grpc-addr"))
		if err != nil {
			log.Err(err).Msg("")
			return err
		}
		authorizer := &authorizer{
			password: c.String("agent-secret"),
		}
		grpcServer := grpc.NewServer(
			grpc.StreamInterceptor(authorizer.streamInterceptor),
			grpc.UnaryInterceptor(authorizer.unaryInterceptor),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime: c.Duration("keepalive-min-time"),
			}),
		)
		woodpeckerServer := woodpeckerGrpcServer.NewWoodpeckerServer(
			_forge,
			server.Config.Services.Queue,
			server.Config.Services.Logs,
			server.Config.Services.Pubsub,
			_store,
			server.Config.Server.Host,
		)
		proto.RegisterWoodpeckerServer(grpcServer, woodpeckerServer)

		err = grpcServer.Serve(lis)
		if err != nil {
			log.Err(err).Msg("")
			return err
		}
		return nil
	})

	proxyWebUI := c.String("www-proxy")
	var webUIServe func(w http.ResponseWriter, r *http.Request)

	if proxyWebUI == "" {
		webUIServe = web.New().ServeHTTP
	} else {
		origin, _ := url.Parse(proxyWebUI)

		director := func(req *http.Request) {
			req.Header.Add("X-Forwarded-Host", req.Host)
			req.Header.Add("X-Origin-Host", origin.Host)
			req.URL.Scheme = origin.Scheme
			req.URL.Host = origin.Host
		}

		proxy := &httputil.ReverseProxy{Director: director}
		webUIServe = proxy.ServeHTTP
	}

	// setup the server and start the listener
	handler := router.Load(
		webUIServe,
		middleware.Logger(time.RFC3339, true),
		middleware.Version,
		middleware.Config(c),
		middleware.Store(c, _store),
	)

	if c.String("server-cert") != "" {
		// start the server with tls enabled
		g.Go(func() error {
			serve := &http.Server{
				Addr:    ":https",
				Handler: handler,
				TLSConfig: &tls.Config{
					NextProtos: []string{"h2", "http/1.1"},
				},
			}
			return serve.ListenAndServeTLS(
				c.String("server-cert"),
				c.String("server-key"),
			)
		})

		// http to https redirect
		redirect := func(w http.ResponseWriter, req *http.Request) {
			serverHost := server.Config.Server.Host
			serverHost = strings.TrimPrefix(serverHost, "http://")
			serverHost = strings.TrimPrefix(serverHost, "https://")
			req.URL.Scheme = "https"
			req.URL.Host = serverHost

			w.Header().Set("Strict-Transport-Security", "max-age=31536000")

			http.Redirect(w, req, req.URL.String(), http.StatusMovedPermanently)
		}

		g.Go(func() error {
			return http.ListenAndServe(":http", http.HandlerFunc(redirect))
		})
	} else if c.Bool("lets-encrypt") {
		// start the server with lets-encrypt
		certmagic.DefaultACME.Email = c.String("lets-encrypt-email")
		certmagic.DefaultACME.Agreed = true

		address, err := url.Parse(c.String("server-host"))
		if err != nil {
			return err
		}

		g.Go(func() error {
			if err := certmagic.HTTPS([]string{address.Host}, handler); err != nil {
				log.Err(err).Msg("certmagic does not work")
				os.Exit(1)
			}
			return nil
		})
	} else {
		// start the server without tls
		g.Go(func() error {
			return http.ListenAndServe(
				c.String("server-addr"),
				handler,
			)
		})
	}

	log.Info().Msgf("Starting Woodpecker server with version '%s'", version.String())

	return g.Wait()
}

func setupEvilGlobals(c *cli.Context, v store.Store, f forge.Forge) {
	// storage
	server.Config.Storage.Files = v

	// forge
	server.Config.Services.Forge = f
	server.Config.Services.Timeout = c.Duration("forge-timeout")

	// services
	server.Config.Services.Queue = setupQueue(c, v)
	server.Config.Services.Logs = logging.New()
	server.Config.Services.Pubsub = pubsub.New()
	if err := server.Config.Services.Pubsub.Create(context.Background(), "topic/events"); err != nil {
		log.Error().Err(err).Msg("could not create pubsub service")
	}
	server.Config.Services.Registries = setupRegistryService(c, v)

	// TODO(1544): fix encrypted store
	// // encryption
	// encryptedSecretStore := encryptedStore.NewSecretStore(v)
	// err := encryption.Encryption(c, v).WithClient(encryptedSecretStore).Build()
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("could not create encryption service")
	// }
	// server.Config.Services.Secrets = setupSecretService(c, encryptedSecretStore)
	server.Config.Services.Secrets = setupSecretService(c, v)

	server.Config.Services.Environ = setupEnvironService(c, v)
	server.Config.Services.Membership = setupMembershipService(c, f)

	server.Config.Services.SignaturePrivateKey, server.Config.Services.SignaturePublicKey = setupSignatureKeys(v)

	if endpoint := c.String("config-service-endpoint"); endpoint != "" {
		server.Config.Services.ConfigService = config.NewHTTP(endpoint, server.Config.Services.SignaturePrivateKey)
	}

	// authentication
	server.Config.Pipeline.AuthenticatePublicRepos = c.Bool("authenticate-public-repos")

	// Cloning
	server.Config.Pipeline.DefaultCloneImage = c.String("default-clone-image")

	// Execution
	_events := c.StringSlice("default-cancel-previous-pipeline-events")
	events := make([]model.WebhookEvent, len(_events))
	for _, v := range _events {
		events = append(events, model.WebhookEvent(v))
	}
	server.Config.Pipeline.DefaultCancelPreviousPipelineEvents = events

	// limits
	server.Config.Pipeline.Limits.MemSwapLimit = c.Int64("limit-mem-swap")
	server.Config.Pipeline.Limits.MemLimit = c.Int64("limit-mem")
	server.Config.Pipeline.Limits.ShmSize = c.Int64("limit-shm-size")
	server.Config.Pipeline.Limits.CPUQuota = c.Int64("limit-cpu-quota")
	server.Config.Pipeline.Limits.CPUShares = c.Int64("limit-cpu-shares")
	server.Config.Pipeline.Limits.CPUSet = c.String("limit-cpu-set")

	// server configuration
	server.Config.Server.Cert = c.String("server-cert")
	server.Config.Server.Key = c.String("server-key")
	server.Config.Server.Pass = c.String("agent-secret")
	server.Config.Server.Host = c.String("server-host")
	if c.IsSet("server-dev-oauth-host") {
		server.Config.Server.OAuthHost = c.String("server-dev-oauth-host")
	} else {
		server.Config.Server.OAuthHost = c.String("server-host")
	}
	server.Config.Server.Port = c.String("server-addr")
	server.Config.Server.Docs = c.String("docs")
	server.Config.Server.StatusContext = c.String("status-context")
	server.Config.Server.StatusContextFormat = c.String("status-context-format")
	server.Config.Server.SessionExpires = c.Duration("session-expires")
	server.Config.Pipeline.Networks = c.StringSlice("network")
	server.Config.Pipeline.Volumes = c.StringSlice("volume")
	server.Config.Pipeline.Privileged = c.StringSlice("escalate")

	// prometheus
	server.Config.Prometheus.AuthToken = c.String("prometheus-auth-token")

	// TODO(485) temporary workaround to not hit api rate limits
	server.Config.FlatPermissions = c.Bool("flat-permissions")
}

type authorizer struct {
	password string
}

func (a *authorizer) streamInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	if err := a.authorize(stream.Context()); err != nil {
		return err
	}
	return handler(srv, stream)
}

func (a *authorizer) unaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if err := a.authorize(ctx); err != nil {
		return nil, err
	}
	return handler(ctx, req)
}

func (a *authorizer) authorize(ctx context.Context) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md["password"]) > 0 && md["password"][0] == a.password {
			return nil
		}
		return errors.New("invalid agent token")
	}
	return errors.New("missing agent token")
}
