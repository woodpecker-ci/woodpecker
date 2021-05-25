package main

import (
	"context"
	"encoding/json"
	"math"
	"net/url"
	"time"

	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/interrupt"
	"github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/rpc"

	_ "github.com/joho/godotenv/autoload"
	"github.com/urfave/cli"
)

var onceCommand = cli.Command{
	Name:   "once",
	Usage:  "execute one build",
	Hidden: false,
	Action: once,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:   "endpoint",
			EnvVar: "PIPED_ENDPOINT,PIPED_SERVER",
			Value:  "ws://localhost:9999",
		},
		cli.StringFlag{
			Name:   "token",
			EnvVar: "PIPED_TOKEN,PIPED_SECRET",
		},
		cli.DurationFlag{
			Name:   "backoff",
			EnvVar: "PIPED_BACKOFF",
			Value:  time.Second * 15,
		},
		cli.IntFlag{
			Name:   "retry-limit",
			EnvVar: "PIPED_RETRY_LIMIT",
			Value:  math.MaxInt32,
		},
		cli.StringFlag{
			Name:   "platform",
			EnvVar: "PIPED_PLATFORM",
			Value:  "linux/amd64",
		},
		cli.StringFlag{
			Name:   "json",
			EnvVar: "PIPED_JSON",
		},
	},
}

func once(c *cli.Context) error {
	endpoint, err := url.Parse(
		c.String("endpoint"),
	)
	if err != nil {
		return err
	}

	client, err := rpc.NewClient(
		endpoint.String(),
		rpc.WithRetryLimit(
			c.Int("retry-limit"),
		),
		rpc.WithBackoff(
			c.Duration("backoff"),
		),
		rpc.WithToken(
			c.String("token"),
		),
	)
	if err != nil {
		return err
	}
	defer client.Close()

	ctx := context.Background()
	ctx = interrupt.WithContextFunc(ctx, func() {
		println("ctrl+c received, terminating process")
	})

	return run(ctx, &onceClient{client, c.String("json")}, rpc.NoFilter)
}

type onceClient struct {
	*rpc.Client
	json string
}

func (c *onceClient) Next(ctx context.Context, filter rpc.Filter) (*rpc.Pipeline, error) {
	in := []byte(c.json)
	out := new(rpc.Pipeline)
	err := json.Unmarshal(in, out)
	return out, err
}
