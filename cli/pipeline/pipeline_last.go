package pipeline

import (
	"os"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var buildLastCmd = &cli.Command{
	Name:      "last",
	Usage:     "show latest build details",
	ArgsUsage: "<repo/name>",
	Action:    buildLast,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplBuildInfo),
		&cli.StringFlag{
			Name:  "branch",
			Usage: "branch name",
			Value: "master",
		},
	),
}

func buildLast(c *cli.Context) error {
	repo := c.Args().First()
	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	build, err := client.PipelineLast(owner, name, c.String("branch"))
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format"))
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, build)
}
