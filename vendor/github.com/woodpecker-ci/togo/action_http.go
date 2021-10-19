package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/urfave/cli"

	"github.com/woodpecker-ci/togo/template"
)

type (
	httpParams struct {
		Encode  bool
		Package string
		Files   []*httpFile
	}
	httpFile struct {
		Base    string
		Name    string
		Path    string
		Ext     string
		Data    string
		Size    int64
		Time    int64
		Encoded bool
	}
)

var httpCommand = cli.Command{
	Name:   "http",
	Usage:  "generate an http filesystem",
	Action: httpAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "package",
			Value: "http",
		},
		cli.StringFlag{
			Name:  "input",
			Value: "files/**",
		},
		cli.StringFlag{
			Name:  "output",
			Value: "http_gen.go",
		},
		cli.StringFlag{
			Name: "exclude",
		},
		cli.StringFlag{
			Name:  "trim-prefix",
			Value: "files",
		},
		cli.StringSliceFlag{
			Name:  "plain-text",
			Value: &cli.StringSlice{"html", "js", "css"},
		},
	},
}

func httpAction(c *cli.Context) error {
	pattern := c.Args().First()
	if pattern == "" {
		pattern = c.String("input")
	}

	fsys := os.DirFS("")
	matches, err := doublestar.Glob(fsys, pattern)
	if err != nil {
		return err
	}

	params := httpParams{
		Encode:  c.Bool("encode"),
		Package: c.String("package"),
	}

	var (
		prefix  = c.String("trim-prefix")
		exclude *regexp.Regexp
	)

	if s := c.String("exclude"); s != "" {
		exclude = regexp.MustCompilePOSIX(s)
	}

	for _, match := range matches {
		stat, oserr := os.Stat(match)
		if oserr != nil {
			return oserr
		}
		if stat.IsDir() {
			continue
		}
		if exclude != nil && exclude.MatchString(match) {
			continue
		}

		raw, ioerr := ioutil.ReadFile(match)
		if ioerr != nil {
			return ioerr
		}
		encoded := true
		switch {
		case strings.HasSuffix(match, ".min.js"):
		case strings.HasSuffix(match, ".min.css"):
		case strings.HasSuffix(match, ".css"):
			encoded = false
		case strings.HasSuffix(match, ".js"):
			encoded = false
		case strings.HasSuffix(match, ".html"):
			encoded = false
		}
		data := string(raw)
		if !encoded {
			data = strings.Replace(data, "`", "`+\"`\"+`", -1)
		}
		params.Files = append(params.Files, &httpFile{
			Path:    strings.TrimPrefix(match, prefix),
			Name:    filepath.Base(match),
			Base:    strings.TrimSuffix(filepath.Base(match), filepath.Ext(match)),
			Ext:     filepath.Ext(match),
			Data:    data,
			Time:    stat.ModTime().Unix(),
			Size:    stat.Size(),
			Encoded: encoded,
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

	return template.Execute(wr, "http.tmpl", &params)
}
