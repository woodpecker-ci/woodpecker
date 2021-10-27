package user

import (
	"fmt"
	"os"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var userInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "show user details",
	ArgsUsage: "<username>",
	Action:    userInfo,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplUserInfo),
	),
}

func userInfo(c *cli.Context) error {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	login := c.Args().First()
	if len(login) == 0 {
		return fmt.Errorf("Missing or invalid user login")
	}

	user, err := client.User(login)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, user)
}

// template for user information
var tmplUserInfo = `User: {{ .Login }}
Email: {{ .Email }}`
