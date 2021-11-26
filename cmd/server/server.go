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
	"path/filepath"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"github.com/woodpecker-ci/woodpecker/pipeline/rpc/proto"
	"github.com/woodpecker-ci/woodpecker/server"
	woodpeckerGrpcServer "github.com/woodpecker-ci/woodpecker/server/grpc"
	"github.com/woodpecker-ci/woodpecker/server/logging"
	"github.com/woodpecker-ci/woodpecker/server/plugins/sender"
	"github.com/woodpecker-ci/woodpecker/server/pubsub"
	"github.com/woodpecker-ci/woodpecker/server/remote"
	"github.com/woodpecker-ci/woodpecker/server/router"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware"
	"github.com/woodpecker-ci/woodpecker/server/store"
	"github.com/woodpecker-ci/woodpecker/server/web"
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

	// debug level if requested by user
	// TODO: format output & options to switch to json aka. option to add channels to send logs to
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
	if c.Bool("debug") {
		log.Warn().Msg("--debug is deprecated, use --log-level instead")
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

	remote_, err := SetupRemote(c)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	store_, err := setupStore(c)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}
	defer func() {
		if err := store_.Close(); err != nil {
			log.Error().Err(err).Msg("could not close store")
		}
	}()

	setupEvilGlobals(c, store_, remote_)

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
		middleware.Store(c, store_),
		middleware.Remote(remote_),
	)

	var g errgroup.Group

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
			grpc.UnaryInterceptor(authorizer.unaryIntercaptor),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime: c.Duration("keepalive-min-time"),
			}),
		)
		woodpeckerServer := woodpeckerGrpcServer.NewWoodpeckerServer(
			remote_,
			server.Config.Services.Queue,
			server.Config.Services.Logs,
			server.Config.Services.Pubsub,
			store_,
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

	setupMetrics(&g, store_)

	// start the server with tls enabled
	if c.String("server-cert") != "" {
		g.Go(func() error {
			return http.ListenAndServe(":http", http.HandlerFunc(redirect))
		})
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
		return g.Wait()
	}

	// start the server without tls enabled
	if !c.Bool("lets-encrypt") {
		return http.ListenAndServe(
			c.String("server-addr"),
			handler,
		)
	}

	// start the server with lets encrypt enabled
	// listen on ports 443 and 80
	address, err := url.Parse(c.String("server-host"))
	if err != nil {
		return err
	}

	dir := cacheDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	manager := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(address.Host),
		Cache:      autocert.DirCache(dir),
	}
	g.Go(func() error {
		return http.ListenAndServe(":http", manager.HTTPHandler(http.HandlerFunc(redirect)))
	})
	g.Go(func() error {
		serve := &http.Server{
			Addr:    ":https",
			Handler: handler,
			TLSConfig: &tls.Config{
				GetCertificate: manager.GetCertificate,
				NextProtos:     []string{"h2", "http/1.1"},
			},
		}
		return serve.ListenAndServeTLS("", "")
	})

	return g.Wait()
}

func setupEvilGlobals(c *cli.Context, v store.Store, r remote.Remote) {
	// storage
	server.Config.Storage.Files = v
	server.Config.Storage.Config = v

	// services
	server.Config.Services.Queue = setupQueue(c, v)
	server.Config.Services.Logs = logging.New()
	server.Config.Services.Pubsub = pubsub.New()
	if err := server.Config.Services.Pubsub.Create(context.Background(), "topic/events"); err != nil {
		log.Error().Err(err).Msg("could not create pubsub service")
	}
	server.Config.Services.Registries = setupRegistryService(c, v)
	server.Config.Services.Secrets = setupSecretService(c, v)
	server.Config.Services.Senders = sender.New(v, v)
	server.Config.Services.Environ = setupEnvironService(c, v)

	if endpoint := c.String("gating-service"); endpoint != "" {
		server.Config.Services.Senders = sender.NewRemote(endpoint)
	}

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
	server.Config.Server.Port = c.String("server-addr")
	server.Config.Server.Docs = c.String("docs")
	server.Config.Server.SessionExpires = c.Duration("session-expires")
	server.Config.Pipeline.Networks = c.StringSlice("network")
	server.Config.Pipeline.Volumes = c.StringSlice("volume")
	server.Config.Pipeline.Privileged = c.StringSlice("escalate")

	// prometheus
	server.Config.Prometheus.AuthToken = c.String("prometheus-auth-token")
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

func (a *authorizer) unaryIntercaptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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

func redirect(w http.ResponseWriter, req *http.Request) {
	serverHost := server.Config.Server.Host
	serverHost = strings.TrimPrefix(serverHost, "http://")
	serverHost = strings.TrimPrefix(serverHost, "https://")
	req.URL.Scheme = "https"
	req.URL.Host = serverHost

	w.Header().Set("Strict-Transport-Security", "max-age=31536000")

	http.Redirect(w, req, req.URL.String(), http.StatusMovedPermanently)
}

func cacheDir() string {
	const base = "golang-autocert"
	if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
		return filepath.Join(xdg, base)
	}
	return filepath.Join(os.Getenv("HOME"), ".cache", base)
}
