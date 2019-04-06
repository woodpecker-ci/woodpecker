package autoscale

import (
	"github.com/urfave/cli"

	"github.com/laszlocph/drone-oss-08/cli/drone/internal"
)

var autoscaleResumeCmd = cli.Command{
	Name:   "resume",
	Usage:  "resume the autoscaler",
	Action: autoscaleResume,
}

func autoscaleResume(c *cli.Context) error {
	client, err := internal.NewAutoscaleClient(c)
	if err != nil {
		return err
	}
	return client.AutoscaleResume()
}
