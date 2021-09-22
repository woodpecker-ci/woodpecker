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
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"

	"github.com/woodpecker-ci/woodpecker/cncd/logging"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/rpc/proto"
	"github.com/woodpecker-ci/woodpecker/cncd/pubsub"
	"github.com/woodpecker-ci/woodpecker/plugins/sender"
	"github.com/woodpecker-ci/woodpecker/remote"
	"github.com/woodpecker-ci/woodpecker/server"
	woodpeckerGrpcServer "github.com/woodpecker-ci/woodpecker/server/grpc"
	"github.com/woodpecker-ci/woodpecker/server/router"
	"github.com/woodpecker-ci/woodpecker/server/router/middleware"
	"github.com/woodpecker-ci/woodpecker/store"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	oldcontext "golang.org/x/net/context"
)

func loop(c *cli.Context) error {

	// debug level if requested by user
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.WarnLevel)
	}

	// must configure the drone_host variable
	if c.String("server-host") == "" {
		logrus.Fatalln("DRONE_HOST/DRONE_SERVER_HOST/WOODPECKER_HOST/WOODPECKER_SERVER_HOST is not properly configured")
	}

	if !strings.Contains(c.String("server-host"), "://") {
		logrus.Fatalln(
			"DRONE_HOST/DRONE_SERVER_HOST/WOODPECKER_HOST/WOODPECKER_SERVER_HOST must be <scheme>://<hostname> format",
		)
	}

	if strings.Contains(c.String("server-host"), "://localhost") {
		logrus.Warningln(
			"DRONE_HOST/DRONE_SERVER_HOST/WOODPECKER_HOST/WOODPECKER_SERVER_HOST should probably be publicly accessible (not localhost)",
		)
	}

	if strings.HasSuffix(c.String("server-host"), "/") {
		logrus.Fatalln(
			"DRONE_HOST/DRONE_SERVER_HOST/WOODPECKER_HOST/WOODPECKER_SERVER_HOST must not have trailing slash",
		)
	}

	remote_, err := SetupRemote(c)
	if err != nil {
		logrus.Fatal(err)
	}

	store_ := setupStore(c)
	setupEvilGlobals(c, store_, remote_)

	// we are switching from gin to httpservermux|treemux,
	// so if this code looks strange, that is why.
	tree := setupTree(c)

	// setup the server and start the listener
	handler := router.Load(
		tree,
		ginrus.Ginrus(logrus.StandardLogger(), time.RFC3339, true),
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
			logrus.Error(err)
			return err
		}
		auther := &authorizer{
			password: c.String("agent-secret"),
		}
		grpcServer := grpc.NewServer(
			grpc.StreamInterceptor(auther.streamInterceptor),
			grpc.UnaryInterceptor(auther.unaryIntercaptor),
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime: c.Duration("keepalive-min-time"),
			}),
		)
		droneServer := woodpeckerGrpcServer.NewDroneServer(
			remote_,
			server.Config.Services.Queue,
			server.Config.Services.Logs,
			server.Config.Services.Pubsub,
			store_,
			server.Config.Server.Host,
		)
		proto.RegisterDroneServer(grpcServer, droneServer)

		err = grpcServer.Serve(lis)
		if err != nil {
			logrus.Error(err)
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
					NextProtos: []string{"http/1.1"}, // disable h2 because Safari :(
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
	os.MkdirAll(dir, 0700)

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
				NextProtos:     []string{"http/1.1"}, // disable h2 because Safari :(
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
	server.Config.Services.Pubsub.Create(context.Background(), "topic/events")
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
	server.Config.Server.RepoConfig = c.String("repo-config")
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

func (a *authorizer) unaryIntercaptor(ctx oldcontext.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
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
	var serverHost string = server.Config.Server.Host
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
