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
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	prometheus_http "github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"golang.org/x/sync/errgroup"

	"go.woodpecker-ci.org/woodpecker/v3/server"
	"go.woodpecker-ci.org/woodpecker/v3/server/cron"
	"go.woodpecker-ci.org/woodpecker/v3/server/router"
	"go.woodpecker-ci.org/woodpecker/v3/server/router/middleware"
	"go.woodpecker-ci.org/woodpecker/v3/server/web"
	"go.woodpecker-ci.org/woodpecker/v3/shared/logger"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

const (
	shutdownTimeout = time.Second * 5
)

var (
	stopServerFunc     context.CancelCauseFunc = func(error) {}
	shutdownCancelFunc context.CancelFunc      = func() {}
	shutdownCtx                                = context.Background()
)

func run(ctx context.Context, c *cli.Command) error {
	if err := logger.SetupGlobalLogger(ctx, c, true); err != nil {
		return err
	}

	ctx, ctxCancel := context.WithCancelCause(ctx)
	stopServerFunc = func(err error) {
		if err != nil {
			log.Error().Err(err).Msg("shutdown of whole server")
		}
		stopServerFunc = func(error) {}
		shutdownCtx, shutdownCancelFunc = context.WithTimeout(shutdownCtx, shutdownTimeout)
		ctxCancel(err)
	}
	defer stopServerFunc(nil)
	defer shutdownCancelFunc()

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

	_store, err := setupStore(ctx, c)
	if err != nil {
		return fmt.Errorf("can't setup store: %w", err)
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

	// wait for all services until one do stops with an error
	serviceWaitingGroup := errgroup.Group{}

	log.Info().Msgf("starting Woodpecker server with version '%s'", version.String())

	serviceWaitingGroup.Go(func() error {
		log.Info().Msg("starting cron service ...")
		if err := cron.Run(ctx, _store); err != nil {
			go stopServerFunc(err)
			return err
		}
		log.Info().Msg("cron service stopped")
		return nil
	})

	// start the grpc server
	serviceWaitingGroup.Go(func() error {
		log.Info().Msg("starting grpc server ...")
		if err := runGrpcServer(ctx, c, _store); err != nil {
			// stop whole server as grpc is essential
			go stopServerFunc(err)
			return err
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

	if c.String("server-cert") != "" {
		// start the server with tls enabled
		serviceWaitingGroup.Go(func() error {
			tlsServer := &http.Server{
				Addr:    server.Config.Server.PortTLS,
				Handler: handler,
				TLSConfig: &tls.Config{
					NextProtos: []string{"h2", "http/1.1"},
				},
			}

			go func() {
				<-ctx.Done()
				log.Info().Msg("shutdown tls server ...")
				if err := tlsServer.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck
					log.Error().Err(err).Msg("shutdown tls server failed")
				} else {
					log.Info().Msg("tls server stopped")
				}
			}()

			log.Info().Msg("starting tls server ...")
			err := tlsServer.ListenAndServeTLS(
				c.String("server-cert"),
				c.String("server-key"),
			)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Msg("TLS server failed")
				stopServerFunc(fmt.Errorf("TLS server failed: %w", err))
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

		serviceWaitingGroup.Go(func() error {
			redirectServer := &http.Server{
				Addr:    server.Config.Server.Port,
				Handler: http.HandlerFunc(redirect),
			}
			go func() {
				<-ctx.Done()
				log.Info().Msg("shutdown redirect server ...")
				if err := redirectServer.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck
					log.Error().Err(err).Msg("shutdown redirect server failed")
				} else {
					log.Info().Msg("redirect server stopped")
				}
			}()

			log.Info().Msg("starting redirect server ...")
			if err := redirectServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Msg("redirect server failed")
				stopServerFunc(fmt.Errorf("redirect server failed: %w", err))
			}
			return nil
		})
	} else {
		// start the server without tls
		serviceWaitingGroup.Go(func() error {
			httpServer := &http.Server{
				Addr:    c.String("server-addr"),
				Handler: handler,
			}

			go func() {
				<-ctx.Done()
				log.Info().Msg("shutdown http server ...")
				if err := httpServer.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck
					log.Error().Err(err).Msg("shutdown http server failed")
				} else {
					log.Info().Msg("http server stopped")
				}
			}()

			log.Info().Msg("starting http server ...")
			if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Msg("http server failed")
				stopServerFunc(fmt.Errorf("http server failed: %w", err))
			}
			return err
		})
	}

	if metricsServerAddr := c.String("metrics-server-addr"); metricsServerAddr != "" {
		startMetricsCollector(ctx, _store)

		serviceWaitingGroup.Go(func() error {
			metricsRouter := gin.New()
			metricsRouter.GET("/metrics", gin.WrapH(prometheus_http.Handler()))

			metricsServer := &http.Server{
				Addr:    metricsServerAddr,
				Handler: metricsRouter,
			}

			go func() {
				<-ctx.Done()
				log.Info().Msg("shutdown metrics server ...")
				if err := metricsServer.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck
					log.Error().Err(err).Msg("shutdown metrics server failed")
				} else {
					log.Info().Msg("metrics server stopped")
				}
			}()

			log.Info().Msg("starting metrics server ...")
			if err := metricsServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Error().Err(err).Msg("metrics server failed")
				stopServerFunc(fmt.Errorf("metrics server failed: %w", err))
			}
			return err
		})
	}

	return serviceWaitingGroup.Wait()
}
