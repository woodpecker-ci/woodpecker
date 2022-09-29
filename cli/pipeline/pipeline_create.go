package pipeline

import (
	"os"
	"strings"
	"text/template"

	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var pipelineCreateCmd = &cli.Command{
	Name:      "create",
	Usage:     "create new pipeline",
	ArgsUsage: "<repo/name>",
	Action:    pipelineCreate,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplPipelineList),
		&cli.StringFlag{
			Name:     "branch",
			Usage:    "branch to create pipeline from",
			Required: true,
		},
		&cli.StringSliceFlag{
			Name:  "var",
			Usage: "key=value",
		},
	),
}

func pipelineCreate(c *cli.Context) error {
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
		sp := strings.SplitN(vaz, "=", 2)
		if len(sp) == 2 {
			variables[sp[0]] = sp[1]
		}
	}

	options := &woodpecker.PipelineOptions{
		Branch:    branch,
		Variables: variables,
	}

	pipeline, err := client.PipelineCreate(owner, name, options)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	if err := tmpl.Execute(os.Stdout, pipeline); err != nil {
		return err
	}

	return nil
}
