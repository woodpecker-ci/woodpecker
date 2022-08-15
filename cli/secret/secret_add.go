package secret

import (
	"io/ioutil"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/woodpecker-ci/woodpecker/cli/common"
	"github.com/woodpecker-ci/woodpecker/cli/internal"
	"github.com/woodpecker-ci/woodpecker/woodpecker-go/woodpecker"
)

var secretCreateCmd = &cli.Command{
	Name:      "add",
	Usage:     "adds a secret",
	ArgsUsage: "[org/repo|org]",
	Action:    secretCreate,
	Flags: append(common.GlobalFlags,
		&cli.BoolFlag{
			Name:  "global",
			Usage: "global secret",
		},
		&cli.StringFlag{
			Name:  "organization",
			Usage: "organizations name (e.g. octocat)",
		},
		&cli.StringFlag{
			Name:  "repository",
			Usage: "repository name (e.g. octocat/hello-world)",
		},
		&cli.StringFlag{
			Name:  "name",
			Usage: "secret name",
		},
		&cli.StringFlag{
			Name:  "value",
			Usage: "secret value",
		},
		&cli.StringSliceFlag{
			Name:  "event",
			Usage: "secret limited to these events",
		},
		&cli.StringSliceFlag{
			Name:  "image",
			Usage: "secret limited to these images",
		},
	),
}

func secretCreate(c *cli.Context) error {
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	secret := &woodpecker.Secret{
		Name:   c.String("name"),
		Value:  c.String("value"),
		Images: c.StringSlice("image"),
		Events: c.StringSlice("event"),
	}
	if len(secret.Events) == 0 {
		secret.Events = defaultSecretEvents
	}
	if strings.HasPrefix(secret.Value, "@") {
		path := strings.TrimPrefix(secret.Value, "@")
		out, ferr := ioutil.ReadFile(path)
		if ferr != nil {
			return ferr
		}
		secret.Value = string(out)
	}
	if c.Bool("global") {
		_, err = client.GlobalSecretCreate(secret)
		return err
	}

	orgName := c.String("organization")
	repoName := c.String("repository")
	if orgName == "" && repoName == "" {
		repoName = c.Args().First()
	}
	if orgName == "" && !strings.Contains(repoName, "/") {
		orgName = repoName
	}
	if orgName != "" {
		_, err = client.OrgSecretCreate(orgName, secret)
		return err
	}

	owner, name, err := internal.ParseRepo(repoName)
	if err != nil {
		return err
	}
	_, err = client.SecretCreate(owner, name, secret)
	return err
}

var defaultSecretEvents = []string{
	woodpecker.EventPush,
	woodpecker.EventTag,
	woodpecker.EventDeploy,
}
