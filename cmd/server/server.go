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
	"net/url"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/gin-gonic/gin"
	prometheus_http "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/cron"
	"go.woodpecker-ci.org/woodpecker/v3/server/store"
	"go.woodpecker-ci.org/woodpecker/v3/shared/logger"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

func run(ctx context.Context, c *cli.Command) error {
	if err := logger.SetupGlobalLogger(ctx, c, true); err != nil {
		return err
	}

	shutdownCtx := context.TODO()

	ctx, ctxCancel := context.WithCancelCause(ctx)
	defer ctxCancel(nil)

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

	// wait for all services until one does stop with an error
	serviceWaitGroup := errgroup.Group{}

	startService := func(name string, startFnc func(context.Context) error, stopFnc func(context.Context) error, stopOnErr bool) {
		if strings.Contains(name, " ") {
			name = fmt.Sprintf("'%s'", name)
		}

		serviceWaitGroup.Go(func() error {
			log.Debug().Msgf("starting %s service ...", name)

			go func() {
				<-ctx.Done()

				log.Debug().Msgf("stopping %s service ...", name)

				if stopFnc != nil {
					if err := stopFnc(shutdownCtx); err != nil {
						log.Error().Err(err).Msgf("failed to stop %s service", name)
						return
					}
				}

				log.Debug().Msgf("%s service stopped", name)
			}()

			if err := startFnc(ctx); err != nil {
				if stopOnErr {
					return err
				}

				log.Error().Err(err).Msgf("could not start %s service", name)
			}

			return nil
		})
	}

	_store, err := backoff.Retry(ctx,
		func() (store.Store, error) {
			return setupStore(ctx, c)
		},
		backoff.WithBackOff(backoff.NewExponentialBackOff()),
		backoff.WithMaxTries(uint(c.Uint("db-max-retries"))),
		backoff.WithNotify(func(err error, delay time.Duration) {
			log.Error().Msgf("failed to setup store: %v: retry in %v", err, delay)
		}))
	if err != nil {
		return err
	}

	defer func() {
		if err := _store.Close(); err != nil {
			log.Error().Err(err).Msg("could not close store")
		}
	}()

	err = setupEvilGlobals(ctx, c, _store)
	if err != nil {
		return fmt.Errorf("can't setup globals: %w", err)
	}

	log.Info().Msgf("starting Woodpecker server with version '%s'", version.String())

	startService("cron",
		func(ctx context.Context) error {
			return cron.Run(ctx, _store)
		},
		nil,
		true,
	)

	startService("grpc",
		func(ctx context.Context) error {
			return runGrpcServer(ctx, c, _store)
		},
		nil,
		true,
	)

	httpHandler, err := getHTTPHandler(c, _store)
	if err != nil {
		return err
	}

	if c.String("server-cert") != "" {
		// start the server with tls enabled
		tlsServer := &http.Server{
			Addr:    server.Config.Server.PortTLS,
			Handler: httpHandler,
			TLSConfig: &tls.Config{
				NextProtos: []string{"h2", "http/1.1"},
			},
		}

		startService("http with tls",
			func(_ context.Context) error {
				log.Info().Msgf("access ui at %s or http://localhost%s", server.Config.Server.Host, server.Config.Server.PortTLS)

				err := tlsServer.ListenAndServeTLS(
					c.String("server-cert"),
					c.String("server-key"),
				)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					return err
				}

				return nil
			},
			func(ctx context.Context) error {
				return tlsServer.Shutdown(ctx)
			},
			true,
		)

		// http to https redirect
		redirect := func(w http.ResponseWriter, req *http.Request) {
			serverURL, _ := url.Parse(server.Config.Server.Host)
			req.URL.Scheme = "https"
			req.URL.Host = serverURL.Host

			w.Header().Set("Strict-Transport-Security", "max-age=31536000")

			http.Redirect(w, req, req.URL.String(), http.StatusMovedPermanently)
		}

		redirectServer := &http.Server{
			Addr:    server.Config.Server.Port,
			Handler: http.HandlerFunc(redirect),
		}

		startService("http redirect",
			func(_ context.Context) error {
				err := redirectServer.ListenAndServe()
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					return err
				}

				return nil
			},
			func(ctx context.Context) error {
				return redirectServer.Shutdown(ctx)
			},
			true,
		)
	} else {
		// start the server without tls
		httpServer := &http.Server{
			Addr:    c.String("server-addr"),
			Handler: httpHandler,
		}

		startService("http",
			func(_ context.Context) error {
				log.Info().Msgf("access ui at %s or http://localhost%s", server.Config.Server.Host, server.Config.Server.Port)

				err := httpServer.ListenAndServe()
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					return err
				}

				return nil
			},
			func(ctx context.Context) error {
				return httpServer.Shutdown(ctx)
			},
			true,
		)
	}

	if metricsServerAddr := c.String("metrics-server-addr"); metricsServerAddr != "" {
		startService("metrics collector", func(ctx context.Context) error {
			startMetricsCollector(ctx, _store)
			return nil
		}, nil, false)

		metricsRouter := gin.New()
		metricsRouter.GET("/metrics", gin.WrapH(prometheus_http.Handler()))

		metricsServer := &http.Server{
			Addr:    metricsServerAddr,
			Handler: metricsRouter,
		}

		startService("metrics",
			func(_ context.Context) error {
				err := metricsServer.ListenAndServe()
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					return err
				}

				return nil
			},
			func(ctx context.Context) error {
				return metricsServer.Shutdown(ctx)
			},
			false,
		)
	}

	return serviceWaitGroup.Wait()
}
