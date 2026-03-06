package main

import (
	"github.com/urfave/cli/v3"

	"go.woodpecker-ci.org/woodpecker/v3/cli/admin"
	"go.woodpecker-ci.org/woodpecker/v3/cli/common"
	"go.woodpecker-ci.org/woodpecker/v3/cli/context"
	"go.woodpecker-ci.org/woodpecker/v3/cli/exec"
	"go.woodpecker-ci.org/woodpecker/v3/cli/info"
	"go.woodpecker-ci.org/woodpecker/v3/cli/lint"
	"go.woodpecker-ci.org/woodpecker/v3/cli/org"
	"go.woodpecker-ci.org/woodpecker/v3/cli/pipeline"
	"go.woodpecker-ci.org/woodpecker/v3/cli/repo"
	"go.woodpecker-ci.org/woodpecker/v3/cli/setup"
	"go.woodpecker-ci.org/woodpecker/v3/cli/update"
	"go.woodpecker-ci.org/woodpecker/v3/version"
)

//go:generate go run docs.go app.go
func newApp() *cli.Command {
	app := &cli.Command{}
	app.Name = "woodpecker-cli"
	app.Description = "Woodpecker command line utility"
	app.Version = version.String()
	app.Usage = "command line utility"
	app.Flags = common.GlobalFlags
	app.Before = common.Before
	app.After = common.After
	app.Suggest = true
	app.ConfigureShellCompletionCommand = func(c *cli.Command) {
		c.Hidden = false
		c.Usage = "generate completion script for the specified shell"
	}
	app.Commands = []*cli.Command{
		admin.Command,
		context.Command,
		exec.Command,
		info.Command,
		lint.Command,
		org.Command,
		pipeline.Command,
		repo.Command,
		setup.Command,
		update.Command,
	}

	return app
}
