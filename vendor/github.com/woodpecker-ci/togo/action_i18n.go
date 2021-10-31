package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli"

	"github.com/woodpecker-ci/togo/template"
)

type (
	i18nParams struct {
		Package string
		Files   []*i18nFile
	}
	i18nFile struct {
		Base string
		Name string
		Path string
		Ext  string
		Data string
	}
)

var i18nCommand = cli.Command{
	Name:   "i18n",
	Usage:  "embed i18n files",
	Action: i18nAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "package",
			Value: "i18n",
		},
		cli.StringFlag{
			Name:  "input",
			Value: "files/*.all.json",
		},
		cli.StringFlag{
			Name:  "output",
			Value: "i18n_gen.go",
		},
		cli.BoolFlag{
			Name: "encode",
		},
	},
}

func i18nAction(c *cli.Context) error {
	pattern := c.Args().First()
	if pattern == "" {
		pattern = c.String("input")
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	params := i18nParams{
		Package: c.String("package"),
	}

	for _, match := range matches {
		raw, ioerr := ioutil.ReadFile(match)
		if ioerr != nil {
			return ioerr
		}
		params.Files = append(params.Files, &i18nFile{
			Path: match,
			Name: filepath.Base(match),
			Base: strings.TrimSuffix(filepath.Base(match), filepath.Ext(match)),
			Ext:  filepath.Ext(match),
			Data: string(raw),
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

	return template.Execute(wr, "i18n.tmpl", params)
}
