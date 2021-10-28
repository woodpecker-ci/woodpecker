package main

import (
	"os"
	"path/filepath"

	"github.com/urfave/cli"

	"github.com/woodpecker-ci/togo/parser"
	"github.com/woodpecker-ci/togo/template"
)

type migration struct {
	Name       string
	Statements []*parser.Statement
}

type logger struct {
	Enabled bool
	Package string
}

type migrationParams struct {
	Package    string
	Dialect    string
	Migrations []migration
	Logger     logger
}

var ddlCommand = cli.Command{
	Name:   "ddl",
	Usage:  "embed ddl statements",
	Action: ddlAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "package",
			Value: "ddl",
		},
		cli.StringFlag{
			Name:  "dialect",
			Value: "sqlite3",
		},
		cli.StringFlag{
			Name:  "input",
			Value: "files/*.sql",
		},
		cli.StringFlag{
			Name:  "output",
			Value: "ddl_gen.go",
		},
		cli.BoolFlag{
			Name: "log",
		},
		cli.StringFlag{
			Name:  "logger",
			Value: "log", // log, logrus
		},
	},
}

func ddlAction(c *cli.Context) error {
	pattern := c.Args().First()
	if pattern == "" {
		pattern = c.String("input")
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	params := migrationParams{
		Package: c.String("package"),
		Dialect: c.String("dialect"),
		Logger: logger{
			Enabled: c.Bool("log"),
			Package: c.String("logger"),
		},
	}

	parse := parser.New()
	for _, match := range matches {
		statements, perr := parse.ParseFile(match)
		if perr != nil {
			return perr
		}
		_, filename := filepath.Split(match)
		params.Migrations = append(params.Migrations, migration{
			Name:       filename,
			Statements: statements,
		})
	}

	wr := os.Stdout
	if output := c.String("output"); output != "" {
		wr, err = os.Create(output)
		if err != nil {
			return err
		}
		defer wr.Close()
	}

	return template.Execute(wr, "ddl.tmpl", params)
}
