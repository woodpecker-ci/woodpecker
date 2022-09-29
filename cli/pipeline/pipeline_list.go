package pipeline

import (
	"os"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var pipelineListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "show pipeline history",
	ArgsUsage: "<repo/name>",
	Action:    pipelineList,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplPipelineList),
		&cli.StringFlag{
			Name:  "branch",
			Usage: "branch filter",
		},
		&cli.StringFlag{
			Name:  "event",
			Usage: "event filter",
		},
		&cli.StringFlag{
			Name:  "status",
			Usage: "status filter",
		},
		&cli.IntFlag{
			Name:  "limit",
			Usage: "limit the list size",
			Value: 25,
		},
	),
}

func pipelineList(c *cli.Context) error {
	repo := c.Args().First()
	owner, name, err := internal.ParseRepo(repo)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	pipelines, err := client.PipelineList(owner, name)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	branch := c.String("branch")
	event := c.String("event")
	status := c.String("status")
	limit := c.Int("limit")

	var count int
	for _, pipeline := range pipelines {
		if count >= limit {
			break
		}
		if branch != "" && pipeline.Branch != branch {
			continue
		}
		if event != "" && pipeline.Event != event {
			continue
		}
		if status != "" && pipeline.Status != status {
			continue
		}
		if err := tmpl.Execute(os.Stdout, pipeline); err != nil {
			return err
		}
		count++
	}
	return nil
}

// template for pipeline list information
var tmplPipelineList = "\x1b[33mBuild #{{ .Number }} \x1b[0m" + `
Status: {{ .Status }}
Event: {{ .Event }}
Commit: {{ .Commit }}
Branch: {{ .Branch }}
Ref: {{ .Ref }}
Author: {{ .Author }} {{ if .Email }}<{{.Email}}>{{ end }}
Message: {{ .Message }}
`
