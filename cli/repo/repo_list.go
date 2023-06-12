package repo

import (
	"os"
	"text/template"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
)

var repoListCmd = &cli.Command{
	Name:      "ls",
	Usage:     "list all repos",
	ArgsUsage: " ",
	Action:    repoList,
	Flags: append(common.GlobalFlags,
		common.FormatFlag(tmplRepoList),
		&cli.StringFlag{
			Name:  "org",
			Usage: "filter by organization",
		},
	),
}

func repoList(c *cli.Context) error {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}

	repos, err := client.RepoList()
	if err != nil || len(repos) == 0 {
		return err
	}

	tmpl, err := template.New("_").Parse(c.String("format") + "\n")
	if err != nil {
		return err
	}

	org := c.String("org")
	for _, repo := range repos {
		if org != "" && org != repo.Owner {
			continue
		}
		if err := tmpl.Execute(os.Stdout, repo); err != nil {
			return err
		}
	}
	return nil
}

// template for repository list items
var tmplRepoList = "\x1b[33m{{ .FullName }}\x1b[0m (id: {{ .ID }})"
