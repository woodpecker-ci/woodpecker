package repo

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

var repoUpdateCmd = &cli.Command{
	Name:      "update",
	Usage:     "update a repository",
	ArgsUsage: "<repo/name>",
	Action:    repoUpdate,
	Flags: append(common.GlobalFlags,
		&cli.BoolFlag{
			Name:  "trusted",
			Usage: "repository is trusted",
		},
		&cli.BoolFlag{
			Name:  "gated",
			Usage: "repository is gated",
		},
		&cli.DurationFlag{
			Name:  "timeout",
			Usage: "repository timeout",
		},
		&cli.StringFlag{
			Name:  "visibility",
			Usage: "repository visibility",
		},
		&cli.StringFlag{
			Name:  "config",
			Usage: "repository configuration path (e.g. .woodpecker.yml)",
		},
		&cli.IntFlag{
			Name:  "pipeline-counter",
			Usage: "repository starting pipeline number",
		},
		&cli.BoolFlag{
			Name:  "unsafe",
			Usage: "validate updating the pipeline-counter is unsafe",
		},
	),
}

func repoUpdate(c *cli.Context) error {
	repo := c.Args().First()
	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	var (
		visibility      = c.String("visibility")
		config          = c.String("config")
		timeout         = c.Duration("timeout")
		trusted         = c.Bool("trusted")
		gated           = c.Bool("gated")
		pipelineCounter = c.Int("pipeline-counter")
		unsafe          = c.Bool("unsafe")
	)

	patch := new(woodpecker.RepoPatch)
	if c.IsSet("trusted") {
		patch.IsTrusted = &trusted
	}
	if c.IsSet("gated") {
		patch.IsGated = &gated
	}
	if c.IsSet("timeout") {
		v := int64(timeout / time.Minute)
		patch.Timeout = &v
	}
	if c.IsSet("config") {
		patch.Config = &config
	}
	if c.IsSet("visibility") {
		switch visibility {
		case "public", "private", "internal":
			patch.Visibility = &visibility
		}
	}
	if c.IsSet("pipeline-counter") && !unsafe {
		fmt.Printf("Setting the pipeline counter is an unsafe operation that could put your repository in an inconsistent state. Please use --unsafe to proceed")
	}
	if c.IsSet("pipeline-counter") && unsafe {
		patch.PipelineCounter = &pipelineCounter
	}

	if _, err := client.RepoPatch(owner, name, patch); err != nil {
		return err
	}
	fmt.Printf("Successfully updated repository %s/%s\n", owner, name)
	return nil
}
