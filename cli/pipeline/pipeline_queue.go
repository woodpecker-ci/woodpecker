package pipeline

import (
	"fmt"
	"os"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var pipelineQueueCmd = &cli.Command{
	Name:      "queue",
	Usage:     "show pipeline queue",
	ArgsUsage: " ",
	Action:    pipelineQueue,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplPipelineQueue),
	),
}

func pipelineQueue(c *cli.Context) error {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	pipelines, err := client.PipelineQueue()
	if err != nil {
		return err
	}

	if len(pipelines) == 0 {
		fmt.Println("there are no pending or running pipelines")
		return nil
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	for _, pipeline := range pipelines {
		if err := tmpl.Execute(os.Stdout, pipeline); err != nil {
			return err
		}
	}
	return nil
}

// template for pipeline list information
var tmplPipelineQueue = "\x1b[33m{{ .FullName }} #{{ .Number }} \x1b[0m" + `
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}
`
