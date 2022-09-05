package build

import (
	"os"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var buildCreateCmd = &cli.Command{
	Name:      "create",
	Usage:     "create new build",
	ArgsUsage: "<repo/name>",
	Action:    buildCreate,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplBuildList),
		&cli.StringFlag{
			Name:     "branch",
			Usage:    "branch to create build from",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "var",
			Usage: "<key>=<value>",
		},
	),
}

func buildCreate(c *cli.Context) error {
	repo := c.Args().First()

	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	branch := c.String("branch")
	variables := make(map[string]string)

	for _, vaz := range c.StringSlice("var") {
		sp := strings.Split(vaz, "=")
		if len(sp) == 2 {
			variables[sp[0]] = sp[1]
		}
	}

	build, err := client.BuildCreate(owner, name, branch, variables)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	if err := tmpl.Execute(os.Stdout, build); err != nil {
		return err
	}

	return nil
}
