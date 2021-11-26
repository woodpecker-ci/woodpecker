package repo

import (
	"os"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var repoInfoCmd = &cli.Command{
	Name:      "info",
	Usage:     "show repository details",
	ArgsUsage: "<repo/name>",
	Action:    repoInfo,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplRepoInfo),
	),
}

func repoInfo(c *cli.Context) error {
	arg := c.Args().First()
	owner, name, err := internal.ParseRepo(arg)
	if err != nil {
		return err
	}

	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	repo, err := client.Repo(owner, name)
	if err != nil {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format"))
	if err != nil {
		return err
	}
	return tmpl.Execute(os.Stdout, repo)
}

// template for repo information
var tmplRepoInfo = `Owner: {{ .Owner }}
Repo: {{ .Name }}
Type: {{ .SCMKind }}
Config: {{ .Config }}
Visibility: {{ .Visibility }}
Private: {{ .IsSCMPrivate }}
Trusted: {{ .IsTrusted }}
Gated: {{ .IsGated }}
Remote: {{ .Clone }}
`
