package user

import (
	"os"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var userListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list all users",
	ArgsUsage: " ",
	Action:    userList,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplUserList),
	),
}

func userList(c *cli.Context) error {
	common.SetupConsoleLogger(c)
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	users, err := client.UserList()
	if err != nil || len(users) == 0 {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}
	for _, user := range users {
		if err := tmpl.Execute(os.Stdout, user); err != nil {
			return err
		}
	}
	return nil
}

// template for user list items
var tmplUserList = `{{ .Login }}`
