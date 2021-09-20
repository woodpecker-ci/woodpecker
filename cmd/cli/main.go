package main

import (
	"fmt"
	"os"

	"github.com/woodpecker-ci/woodpecker/cli/build"
	"github.com/woodpecker-ci/woodpecker/cli/deploy"
	"github.com/woodpecker-ci/woodpecker/cli/exec"
	"github.com/woodpecker-ci/woodpecker/cli/info"
	"github.com/woodpecker-ci/woodpecker/cli/log"
	"github.com/woodpecker-ci/woodpecker/cli/registry"
	"github.com/woodpecker-ci/woodpecker/cli/repo"
	"github.com/woodpecker-ci/woodpecker/cli/secret"
	"github.com/woodpecker-ci/woodpecker/cli/user"
	"github.com/woodpecker-ci/woodpecker/version"

	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "woodpecker"
	app.Version = version.String()
	app.Usage = "command line utility"
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "t, token",
			Usage:  "server auth token",
			EnvVar: "WOODPECKER_TOKEN",
		},

		cli.StringFlag{
			EnvVar: "WOODPECKER_SERVER",
			Name:   "s, server",
			Usage:  "server address",
		},
		cli.BoolFlag{
			EnvVar: "WOODPECKER_SKIP_VERIFY",
			Name:   "skip-verify",
			Usage:  "skip ssl verification",
			Hidden: true,
		},
		cli.StringFlag{
			EnvVar: "SOCKS_PROXY",
			Name:   "socks-proxy",
			Usage:  "socks proxy address",
			Hidden: true,
		},
		cli.BoolFlag{
			EnvVar: "SOCKS_PROXY_OFF",
			Name:   "socks-proxy-off",
			Usage:  "socks proxy ignored",
			Hidden: true,
		},
	}
	app.Commands = []cli.Command{
		build.Command,
		log.Command,
		deploy.Command,
		exec.Command,
		info.Command,
		registry.Command,
		secret.Command,
		repo.Command,
		user.Command,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
