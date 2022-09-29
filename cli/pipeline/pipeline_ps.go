package pipeline

import (
	"os"
	"strconv"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var pipelinePsCmd = &cli.Command{
	Name:      "ps",
	Usage:     "show pipeline steps",
	ArgsUsage: "<repo/name> [pipeline]",
	Action:    pipelinePs,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplPipelinePs),
	),
}

func pipelinePs(c *cli.Context) error {
	repo := c.Args().First()

	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	pipelineArg := c.Args().Get(1)
	var number int

	if pipelineArg == "last" || len(pipelineArg) == 0 {
		// Fetch the pipeline number from the last pipeline
		pipeline, err := client.PipelineLast(owner, name, "")
		if err != nil {
			return err
		}

		number = pipeline.Number
	} else {
		number, err = strconv.Atoi(pipelineArg)
		if err != nil {
			return err
		}
	}

	pipeline, err := client.Pipeline(owner, name, number)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	for _, proc := range pipeline.Procs {
		for _, child := range proc.Children {
			if err := tmpl.Execute(os.Stdout, child); err != nil {
				return err
			}
		}
	}

	return nil
}

// template for pipeline ps information
var tmplPipelinePs = "\x1b[33mProc #{{ .PID }} \x1b[0m" + `
Step: {{ .Name }}
State: {{ .State }}
`
