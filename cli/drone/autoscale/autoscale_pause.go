package autoscale

import (
	"github.com/urfave/cli"

	"github.com/laszlocph/drone-oss-08/cli/drone/internal"
)

var autoscalePauseCmd = cli.Command{
	Name:   "pause",
	Usage:  "pause the autoscaler",
	Action: autoscalePause,
}

func autoscalePause(c *cli.Context) error {
	client, err := internal.NewAutoscaleClient(c)
	if err != nil {
		return err
	}
	return client.AutoscalePause()
}
