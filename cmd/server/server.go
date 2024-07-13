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
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/gin-gonic/gin"
	prometheus_http "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"

	"go.woodpecker-ci.org/woodpecker/v2/pipeline/rpc/proto"
	"go.woodpecker-ci.org/woodpecker/v2/server"
	"go.woodpecker-ci.org/woodpecker/v2/server/cron"
	woodpeckerGrpcServer "go.woodpecker-ci.org/woodpecker/v2/server/grpc"
	"go.woodpecker-ci.org/woodpecker/v2/server/router"
	"go.woodpecker-ci.org/woodpecker/v2/server/router/middleware"
	"go.woodpecker-ci.org/woodpecker/v2/server/web"
	"go.woodpecker-ci.org/woodpecker/v2/shared/logger"
	"go.woodpecker-ci.org/woodpecker/v2/version"
)

func run(c *cli.Context) error {
	if err := logger.SetupGlobalLogger(c, true); err != nil {
		return err
	}

	// set gin mode based on log level
	if zerolog.GlobalLevel() > zerolog.DebugLevel {
		gin.SetMode(gin.ReleaseMode)
	}

	if c.String("server-host") == "" {
		return fmt.Errorf("WOODPECKER_HOST is not properly configured")
	}

	if !strings.Contains(c.String("server-host"), "://") {
		return fmt.Errorf("WOODPECKER_HOST must be <scheme>://<hostname> format")
	}

	if _, err := url.Parse(c.String("server-host")); err != nil {
		return fmt.Errorf("could not parse WOODPECKER_HOST: %w", err)
	}

	if strings.Contains(c.String("server-host"), "://localhost") {
		log.Warn().Msg(
			"WOODPECKER_HOST should probably be publicly accessible (not localhost)",
		)
	}

	_store, err := setupStore(c)
	if err != nil {
		return fmt.Errorf("can't setup store: %w", err)
	}
	defer func() {
		if err := _store.Close(); err != nil {
			log.Error().Err(err).Msg("could not close store")
		}
	}()

	err = setupEvilGlobals(c, _store)
	if err != nil {
		return fmt.Errorf("can't setup globals: %w", err)
	}

	var g errgroup.Group

	setupMetrics(&g, _store)

	g.Go(func() error {
		return cron.Start(c.Context, _store)
	})

	// start the grpc server
	g.Go(func() error {
		lis, err := net.Listen("tcp", c.String("grpc-addr"))
		if err != nil {
			log.Fatal().Err(err).Msg("failed to listen on grpc-addr") //nolint:forbidigo
		}

		jwtSecret := c.String("grpc-secret")
		jwtManager := woodpeckerGrpcServer.NewJWTManager(jwtSecret)

		authorizer := woodpeckerGrpcServer.NewAuthorizer(jwtManager)
		grpcServer := grpc.NewServer(
			grpc.StreamInterceptor(authorizer.StreamInterceptor),
			grpc.UnaryInterceptor(authorizer.UnaryInterceptor),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime: c.Duration("keepalive-min-time"),
			}),
		)

		woodpeckerServer := woodpeckerGrpcServer.NewWoodpeckerServer(
			server.Config.Services.Queue,
			server.Config.Services.Logs,
			server.Config.Services.Pubsub,
			_store,
		)
		proto.RegisterWoodpeckerServer(grpcServer, woodpeckerServer)

		woodpeckerAuthServer := woodpeckerGrpcServer.NewWoodpeckerAuthServer(
			jwtManager,
			server.Config.Server.AgentToken,
			_store,
		)
		proto.RegisterWoodpeckerAuthServer(grpcServer, woodpeckerAuthServer)

		err = grpcServer.Serve(lis)
		if err != nil {
			log.Fatal().Err(err).Msg("failed to serve grpc server") //nolint:forbidigo
		}
		return nil
	})

	proxyWebUI := c.String("www-proxy")
	var webUIServe func(w http.ResponseWriter, r *http.Request)

	if proxyWebUI == "" {
		webEngine, err := web.New()
		if err != nil {
			log.Error().Err(err).Msg("failed to create web engine")
			return err
		}
		webUIServe = webEngine.ServeHTTP
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
		middleware.Store(_store),
	)

	switch {
	case c.String("server-cert") != "":
		// start the server with tls enabled
		g.Go(func() error {
			serve := &http.Server{
				Addr:    server.Config.Server.PortTLS,
				Handler: handler,
				TLSConfig: &tls.Config{
					NextProtos: []string{"h2", "http/1.1"},
				},
			}
			err = serve.ListenAndServeTLS(
				c.String("server-cert"),
				c.String("server-key"),
			)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal().Err(err).Msg("failed to start server with tls") //nolint:forbidigo
			}
			return err
		})

		// http to https redirect
		redirect := func(w http.ResponseWriter, req *http.Request) {
			serverURL, _ := url.Parse(server.Config.Server.Host)
			req.URL.Scheme = "https"
			req.URL.Host = serverURL.Host

			w.Header().Set("Strict-Transport-Security", "max-age=31536000")

			http.Redirect(w, req, req.URL.String(), http.StatusMovedPermanently)
		}

		g.Go(func() error {
			err := http.ListenAndServe(server.Config.Server.Port, http.HandlerFunc(redirect))
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal().Err(err).Msg("unable to start server to redirect from http to https") //nolint:forbidigo
			}
			return err
		})
	case c.Bool("lets-encrypt"):
		// start the server with lets-encrypt
		certmagic.DefaultACME.Email = c.String("lets-encrypt-email")
		certmagic.DefaultACME.Agreed = true

		address, err := url.Parse(strings.TrimSuffix(c.String("server-host"), "/"))
		if err != nil {
			return err
		}

		g.Go(func() error {
			if err := certmagic.HTTPS([]string{address.Host}, handler); err != nil {
				log.Fatal().Err(err).Msg("certmagic does not work") //nolint:forbidigo
			}
			return nil
		})
	default:
		// start the server without tls
		g.Go(func() error {
			err := http.ListenAndServe(
				c.String("server-addr"),
				handler,
			)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal().Err(err).Msg("could not start server") //nolint:forbidigo
			}
			return err
		})
	}

	if metricsServerAddr := c.String("metrics-server-addr"); metricsServerAddr != "" {
		g.Go(func() error {
			metricsRouter := gin.New()
			metricsRouter.GET("/metrics", gin.WrapH(prometheus_http.Handler()))
			err := http.ListenAndServe(metricsServerAddr, metricsRouter)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatal().Err(err).Msg("could not start metrics server") //nolint:forbidigo
			}
			return err
		})
	}

	log.Info().Msgf("starting Woodpecker server with version '%s'", version.String())

	return g.Wait()
}
