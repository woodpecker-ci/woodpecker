package main

import (
	"fmt"

	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/frontend/yaml"
	"github.com/laszlocph/drone-oss-08/cncd/pipeline/pipeline/frontend/yaml/linter"

	"github.com/kr/pretty"
	"github.com/urfave/cli"
)

var lintCommand = cli.Command{
	Name:   "lint",
	Usage:  "lints the yaml file",
	Action: lintAction,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name: "trusted",
		},
		cli.BoolFlag{
			Name: "pretty",
		},
	},
}

func lintAction(c *cli.Context) error {
	file := c.Args().First()
	if file == "" {
		return fmt.Errorf("Error: please provide a path the configuration file")
	}

	conf, err := yaml.ParseFile(file)
	if err != nil {
		return err
	}

	err = linter.New(
		linter.WithTrusted(
			c.Bool("trusted"),
		),
	).Lint(conf)

	if err != nil {
		return err
	}

	if c.Bool("pretty") {
		pretty.Println(conf)
	}

	fmt.Println("Lint complete. Yaml file is valid")
	return nil
}
