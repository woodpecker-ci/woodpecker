package internal

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"golang.org/x/net/proxy"
	"golang.org/x/oauth2"

	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
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
		return nil, fmt.Errorf("you must provide the Woodpecker server address")
	}
	if len(token) == 0 {
		return nil, fmt.Errorf("you must provide your Woodpecker access token")
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
func ParseRepo(str string) (user, repo string, err error) {
	parts := strings.Split(str, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("invalid or missing repository. eg octocat/hello-world")
		return
	}
	user = parts[0]
	repo = parts[1]
	return
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
