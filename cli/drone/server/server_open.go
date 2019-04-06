package server

import (
	"fmt"
	"net/url"

	"github.com/pkg/browser"
	"github.com/urfave/cli"

	"github.com/laszlocph/drone-oss-08/cli/drone/internal"
)

//
// support for cadvisor was temporarily disabled, so
// this command has been hidden from the --help menu
// until available.
//

var serverOpenCmd = cli.Command{
	Name:      "open",
	Usage:     "open server dashboard",
	ArgsUsage: "<servername>",
	Action:    serverOpen,
	Hidden:    true,
}

func serverOpen(c *cli.Context) error {
	client, err := internal.NewAutoscaleClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if len(name) == 0 {
		return fmt.Errorf("Missing or invalid server name")
	}

	server, err := client.Server(name)
	if err != nil {
		return err
	}

	uri := new(url.URL)
	uri.Scheme = "http"
	uri.Host = server.Address + ":8080"
	uri.User = url.UserPassword("admin", server.Secret)

	return browser.OpenURL(uri.String())
}
