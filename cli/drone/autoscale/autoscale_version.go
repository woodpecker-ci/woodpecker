package autoscale

import (
	"os"
	"text/template"

	"github.com/urfave/cli"

	"github.com/laszlocph/drone-oss-08/cli/drone/internal"
)

var autoscaleVersionCmd = cli.Command{
	Name:   "version",
	Usage:  "server version",
	Action: autoscaleVersion,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "format",
			Usage:  "format output",
			Value:  tmplAutoscaleVersion,
			Hidden: true,
		},
	},
}

func autoscaleVersion(c *cli.Context) error {
	client, err := internal.NewAutoscaleClient(c)
	if err != nil {
		return err
	}

	version, err := client.AutoscaleVersion()
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, version)
}

var tmplAutoscaleVersion = `Version: {{ .Version }}
Commit: {{ .Commit }}
Source: {{ .Source }}
`
