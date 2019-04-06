package server

import (
	"os"
	"text/template"

	"github.com/urfave/cli"

	"github.com/laszlocph/drone-oss-08/cli/drone/internal"
)

var serverCreateCmd = cli.Command{
	Name:   "create",
	Usage:  "crate a new server",
	Action: serverCreate,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "format",
			Usage:  "format output",
			Value:  tmplServerCreate,
			Hidden: true,
		},
	},
}

func serverCreate(c *cli.Context) error {
	client, err := internal.NewAutoscaleClient(c)
	if err != nil {
		return err
	}

	server, err := client.ServerCreate()
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, server)
}

var tmplServerCreate = `Name: {{ .Name }}
State: {{ .State }}
`
