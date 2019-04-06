package server

import (
	"fmt"
	"os"
	"text/template"

	"github.com/urfave/cli"

	"github.com/laszlocph/drone-oss-08/cli/drone/internal"
)

var serverInfoCmd = cli.Command{
	Name:      "info",
	Usage:     "show server details",
	ArgsUsage: "<servername>",
	Action:    serverInfo,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "format",
			Usage:  "format output",
			Value:  tmplServerInfo,
			Hidden: true,
		},
	},
}

func serverInfo(c *cli.Context) error {
	client, err := internal.NewAutoscaleClient(c)
	if err != nil {
		return err
	}

	name := c.Args().First()
	if len(name) == 0 {
		return fmt.Errorf("Missing or invalid server name")
	}

	server, err := client.Server(name)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, server)
}

// template for server information
var tmplServerInfo = `Name: {{ .Name }}
Address: {{ .Address }}
Region:  {{ .Region }}
Size:    {{.Size}}
State:   {{ .State }}
{{ if .Error -}}
Error:   {{ .Error }}
{{ end -}}
`
