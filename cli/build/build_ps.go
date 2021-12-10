package build

import (
	"os"
	"strconv"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var buildPsCmd = &cli.Command{
	Name:      "ps",
	Usage:     "show build steps",
	ArgsUsage: "<repo/name> [build]",
	Action:    buildPs,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplBuildPs),
	),
}

func buildPs(c *cli.Context) error {
	repo := c.Args().First()

	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	buildArg := c.Args().Get(1)
	var number int

	if buildArg == "last" || len(buildArg) == 0 {
		// Fetch the build number from the last build
		build, err := client.BuildLast(owner, name, "")
		if err != nil {
			return err
		}

		number = build.Number
	} else {
		number, err = strconv.Atoi(buildArg)
		if err != nil {
			return err
		}
	}

	build, err := client.Build(owner, name, number)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	for _, proc := range build.Procs {
		for _, child := range proc.Children {
			if err := tmpl.Execute(os.Stdout, child); err != nil {
				return err
			}
		}
	}

	return nil
}

// template for build ps information
var tmplBuildPs = "\x1b[33mProc #{{ .PID }} \x1b[0m" + `
Step: {{ .Name }}
State: {{ .State }}
`
