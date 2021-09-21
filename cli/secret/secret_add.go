package secret

import (
	"io/ioutil"
	"strings"

	"github.com/woodpecker-ci/woodpecker/cli/internal"
	"github.com/woodpecker-ci/woodpecker/drone-go/drone"

	"github.com/urfave/cli"
)

var secretCreateCmd = cli.Command{
	Name:      "add",
	Usage:     "adds a secret",
	ArgsUsage: "[repo/name]",
	Action:    secretCreate,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "repository",
			Usage: "repository name (e.g. octocat/hello-world)",
		},
		cli.StringFlag{
			Name:  "name",
			Usage: "secret name",
		},
		cli.StringFlag{
			Name:  "value",
			Usage: "secret value",
		},
		cli.StringSliceFlag{
			Name:  "event",
			Usage: "secret limited to these events",
		},
		cli.StringSliceFlag{
			Name:  "image",
			Usage: "secret limited to these images",
		},
	},
}

func secretCreate(c *cli.Context) error {
	reponame := c.String("repository")
	if reponame == "" {
		reponame = c.Args().First()
	}
	owner, name, err := internal.ParseRepo(reponame)
	if err != nil {
		return err
	}
	client, err := internal.NewClient(c)
	if err != nil {
		return err
	}
	secret := &drone.Secret{
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
	_, err = client.SecretCreate(owner, name, secret)
	return err
}

var defaultSecretEvents = []string{
	drone.EventPush,
	drone.EventTag,
	drone.EventDeploy,
}
