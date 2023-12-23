// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package internal

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/proxy"
	"golang.org/x/oauth2"

	"go.woodpecker-ci.org/woodpecker/v2/woodpecker-go/woodpecker"
)

// NewClient returns a new client from the CLI context.
func NewClient(c *cli.Context) (woodpecker.Client, error) {
	var (
		skip     = c.Bool("skip-verify")
		socks    = c.String("socks-proxy")
		socksoff = c.Bool("socks-proxy-off")
		token    = c.String("token")
		server   = c.String("server")
	)
	server = strings.TrimRight(server, "/")

	// if no server url is provided we can default
	// to the hosted Woodpecker service.
	if len(server) == 0 {
		return nil, fmt.Errorf("Error: you must provide the Woodpecker server address")
	}
	if len(token) == 0 {
		return nil, fmt.Errorf("Error: you must provide your Woodpecker access token")
	}

	// attempt to find system CA certs
	certs, err := x509.SystemCertPool()
	if err != nil {
		log.Error().Msgf("failed to find system CA certs: %v", err)
	}
	tlsConfig := &tls.Config{
		RootCAs:            certs,
		InsecureSkipVerify: skip,
	}

	config := new(oauth2.Config)
	client := config.Client(
		c.Context,
		&oauth2.Token{
			AccessToken: token,
		},
	)

	trans, _ := client.Transport.(*oauth2.Transport)

	if len(socks) != 0 && !socksoff {
		dialer, err := proxy.SOCKS5("tcp", socks, nil, proxy.Direct)
		if err != nil {
			return nil, err
		}
		trans.Base = &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
			Dial:            dialer.Dial,
		}
	} else {
		trans.Base = &http.Transport{
			TLSClientConfig: tlsConfig,
			Proxy:           http.ProxyFromEnvironment,
		}
	}

	return woodpecker.NewClient(server, client), nil
}

// ParseRepo parses the repository owner and name from a string.
func ParseRepo(client woodpecker.Client, str string) (repoID int64, err error) {
	if strings.Contains(str, "/") {
		repo, err := client.RepoLookup(str)
		if err != nil {
			return 0, err
		}
		return repo.ID, nil
	}

	return strconv.ParseInt(str, 10, 64)
}

// ParseKeyPair parses a key=value pair.
func ParseKeyPair(p []string) map[string]string {
	params := map[string]string{}
	for _, i := range p {
		parts := strings.SplitN(i, "=", 2)
		if len(parts) != 2 {
			continue
		}
		params[parts[0]] = parts[1]
	}
	return params
}
