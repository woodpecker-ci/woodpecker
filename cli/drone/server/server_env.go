package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"text/template"

	"github.com/urfave/cli"

	"github.com/laszlocph/drone-oss-08/cli/drone/internal"
	"github.com/drone/drone-go/drone"
)

var serverEnvCmd = cli.Command{
	Name:      "env",
	ArgsUsage: "<servername>",
	Action:    serverEnv,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "shell",
			Usage: "shell [bash, fish, powershell]",
			Value: "bash",
		},
		cli.BoolFlag{
			Name:  "no-proxy",
			Usage: "configure the noproxy variable",
		},
		cli.BoolFlag{
			Name:  "clear",
			Usage: "clear the certificate cache",
		},
	},
}

func serverEnv(c *cli.Context) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	name := c.Args().First()
	if len(name) == 0 {
		return fmt.Errorf("Missing or invalid server name")
	}

	home := path.Join(u.HomeDir, ".drone", "certs")
	base := path.Join(home, name)

	if c.Bool("clean") {
		os.RemoveAll(home)
	}

	server := new(drone.Server)
	if _, err := os.Stat(base); err == nil {
		data, err := ioutil.ReadFile(path.Join(base, "server.json"))
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, server)
		if err != nil {
			return err
		}
	} else {
		client, err := internal.NewAutoscaleClient(c)
		if err != nil {
			return err
		}
		server, err = client.Server(name)
		if err != nil {
			return err
		}
		data, err := json.Marshal(server)
		if err != nil {
			return err
		}
		err = os.MkdirAll(base, 0755)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(path.Join(base, "server.json"), data, 0644)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(path.Join(base, "ca.pem"), server.CACert, 0644)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(path.Join(base, "cert.pem"), server.TLSCert, 0644)
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(path.Join(base, "key.pem"), server.TLSKey, 0644)
		if err != nil {
			return err
		}
	}

	return shellT.Execute(os.Stdout, map[string]interface{}{
		"Name":    server.Name,
		"Address": server.Address,
		"Path":    base,
		"Shell":   c.String("shell"),
		"NoProxy": c.Bool("no-proxy"),
	})
}

var shellT = template.Must(template.New("_").Parse(`
{{- if eq .Shell "fish" -}}
sex -x DOCKER_TLS "1";
set -x DOCKER_TLS_VERIFY "";
set -x DOCKER_CERT_PATH {{ printf "%q" .Path }};
set -x DOCKER_HOST "tcp://{{ .Address }}:2376";
{{ if .NoProxy -}}
set -x NO_PROXY {{ printf "%q" .Address }};
{{ end }}
# Run this command to configure your shell:
# eval "$(drone server env {{ .Name }} --shell=fish)"
{{- else if eq .Shell "powershell" -}}
$Env:DOCKER_TLS = "1"
$Env:DOCKER_TLS_VERIFY = ""
$Env:DOCKER_CERT_PATH = {{ printf "%q" .Path }}
$Env:DOCKER_HOST = "tcp://{{ .Address }}:2376"
{{ if .NoProxy -}}
$Env:NO_PROXY = {{ printf "%q" .Address }}
{{ end }}
# Run this command to configure your shell:
# drone server env {{ .Name }} --shell=powershell | Invoke-Expression
{{- else -}}
export DOCKER_TLS=1
export DOCKER_TLS_VERIFY=
export DOCKER_CERT_PATH={{ .Path }}
export DOCKER_HOST=tcp://{{ .Address }}:2376
{{ if .NoProxy -}}
export NO_PROXY={{ .Address }}
{{ end }}
# Run this command to configure your shell:
# eval "$(drone server env {{ .Name }})"
{{- end }}
`))
