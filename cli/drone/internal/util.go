package internal

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackspirou/syscerts"
	"github.com/urfave/cli"
	"golang.org/x/net/proxy"
	"golang.org/x/oauth2"

	"github.com/drone/drone-go/drone"
)

// NewClient returns a new client from the CLI context.
func NewClient(c *cli.Context) (drone.Client, error) {
	var (
		skip     = c.GlobalBool("skip-verify")
		socks    = c.GlobalString("socks-proxy")
		socksoff = c.GlobalBool("socks-proxy-off")
		token    = c.GlobalString("token")
		server   = c.GlobalString("server")
	)
	server = strings.TrimRight(server, "/")

	// if no server url is provided we can default
	// to the hosted Drone service.
	if len(server) == 0 {
		return nil, fmt.Errorf("Error: you must provide the Drone server address.")
	}
	if len(token) == 0 {
		return nil, fmt.Errorf("Error: you must provide your Drone access token.")
	}

	// attempt to find system CA certs
	certs := syscerts.SystemRootsPool()
	tlsConfig := &tls.Config{
		RootCAs:            certs,
		InsecureSkipVerify: skip,
	}

	config := new(oauth2.Config)
	auther := config.Client(
		oauth2.NoContext,
		&oauth2.Token{
			AccessToken: token,
		},
	)

	trans, _ := auther.Transport.(*oauth2.Transport)

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

	return drone.NewClient(server, auther), nil
}

// NewAutoscaleClient returns a new client from the CLI context.
func NewAutoscaleClient(c *cli.Context) (drone.Client, error) {
	client, err := NewClient(c)
	if err != nil {
		return nil, err
	}
	autoscaler := c.GlobalString("autoscaler")
	if autoscaler == "" {
		return nil, fmt.Errorf("Please provide the autoscaler address")
	}
	client.SetAddress(
		strings.TrimSuffix(autoscaler, "/"),
	)
	return client, nil
}

// ParseRepo parses the repository owner and name from a string.
func ParseRepo(str string) (user, repo string, err error) {
	var parts = strings.Split(str, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("Error: Invalid or missing repository. eg octocat/hello-world.")
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
