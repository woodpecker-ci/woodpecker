package secret

import (
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
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

func parseTargetArgs(client woodpecker.Client, c *cli.Context) (global bool, owner string, repoID int64, err error) {
	if c.Bool("global") {
		return true, "", -1, nil
	}

	repoIDOrFullName := c.String("repository")
	if repoIDOrFullName == "" {
		repoIDOrFullName = c.Args().First()
	}

	orgName := c.String("organization")
	if orgName != "" && repoIDOrFullName == "" {
		return false, orgName, -1, err
	}

	if orgName != "" && !strings.Contains(repoIDOrFullName, "/") {
		repoIDOrFullName = orgName + "/" + repoIDOrFullName
	}

	repoID, err = internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return false, "", -1, err
	}

	return false, "", repoID, nil
}
