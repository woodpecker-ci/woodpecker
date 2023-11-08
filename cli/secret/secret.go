package secret

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/urfave/cli/v2"

	"go.woodpecker-ci.org/woodpecker/cli/common"
	"go.woodpecker-ci.org/woodpecker/cli/internal"
	"go.woodpecker-ci.org/woodpecker/woodpecker-go/woodpecker"
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

func parseTargetArgs(client woodpecker.Client, c *cli.Context) (global bool, orgID, repoID int64, err error) {
	if c.Bool("global") {
		return true, -1, -1, nil
	}

	repoIDOrFullName := c.String("repository")
	if repoIDOrFullName == "" {
		repoIDOrFullName = c.Args().First()
	}

	orgIDOrName := c.String("organization")
	if orgIDOrName == "" && repoIDOrFullName == "" {
		if err := cli.ShowSubcommandHelp(c); err != nil {
			return false, -1, -1, err
		}

		return false, -1, -1, fmt.Errorf("missing arguments")
	}

	if orgIDOrName != "" && repoIDOrFullName == "" {
		if orgID, err := strconv.ParseInt(orgIDOrName, 10, 64); err == nil {
			return false, orgID, -1, nil
		}

		org, err := client.OrgLookup(orgIDOrName)
		if err != nil {
			return false, -1, -1, err
		}

		return false, org.ID, -1, nil
	}

	if orgIDOrName != "" && !strings.Contains(repoIDOrFullName, "/") {
		repoIDOrFullName = orgIDOrName + "/" + repoIDOrFullName
	}

	repoID, err = internal.ParseRepo(client, repoIDOrFullName)
	if err != nil {
		return false, -1, -1, err
	}

	return false, -1, repoID, nil
}
