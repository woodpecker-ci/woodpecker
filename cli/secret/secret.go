package secret

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

// Command exports the secret command.
var Command = &cli.Command{
	Name:  "secret",
	Usage: "manage secrets",
	Flags: common.GlobalFlags,
	Subcommands: []*cli.Command{
		secretCreateCmd,
		secretDeleteCmd,
		secretUpdateCmd,
		secretInfoCmd,
		secretListCmd,
	},
}

func parseTargetArgs(c *cli.Context) (global bool, owner, name string, err error) {
	if c.Bool("global") {
		return true, "", "", nil
	}
	orgName := c.String("organization")
	repoName := c.String("repository")
	if orgName == "" && repoName == "" {
		repoName = c.Args().First()
	}
	if orgName == "" && !strings.Contains(repoName, "/") {
		orgName = repoName
	}
	if orgName != "" {
		return false, orgName, "", err
	}
	owner, name, err = internal.ParseRepo(repoName)
	if err != nil {
		return false, "", "", err
	}
	return false, owner, name, nil
}
