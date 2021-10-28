package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/textproto"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/urfave/cli"

	"github.com/woodpecker-ci/togo/template"
)

type (
	httptestParams struct {
		Package string
		Routes  []*httptestRoute
	}
	httptestRoute struct {
		Source string
		Method string
		Path   string
		Body   string
		Status int
		Header map[string]string
	}
)

var httptestCommand = cli.Command{
	Name:   "httptest",
	Usage:  "generate httptest server",
	Action: httptestAction,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "package",
			Value: "testdata",
		},
		cli.StringFlag{
			Name:  "input",
			Value: "files/**",
		},
		cli.StringFlag{
			Name:  "output",
			Value: "testdata_gen.go",
		},
		cli.StringFlag{
			Name: "exclude",
		},
	},
}

func httptestAction(c *cli.Context) error {
	pattern := c.Args().First()
	if pattern == "" {
		pattern = c.String("input")
	}
	fsys := os.DirFS("")
	matches, err := doublestar.Glob(fsys, pattern)
	if err != nil {
		return err
	}

	var exclude *regexp.Regexp
	if s := c.String("exclude"); s != "" {
		exclude = regexp.MustCompilePOSIX(s)
	}

	params := httptestParams{
		Package: c.String("package"),
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

		log.Printf("parsing file %s", match)

		raw, ioerr := ioutil.ReadFile(match)
		if ioerr != nil {
			return ioerr
		}

		route, parseErr := parseRoute(raw)
		if parseErr != nil {
			return parseErr
		}
		route.Source = strings.TrimPrefix(match, "files/")

		params.Routes = append(params.Routes, route)
	}

	wr := os.Stdout
	if output := c.String("output"); output != "" {
		wr, err = os.Create(output)
		if err != nil {
			return err
		}
		defer wr.Close()
	}

	return template.Execute(wr, "httptest.tmpl", &params)
}

func parseRoute(in []byte) (*httptestRoute, error) {
	out := new(httptestRoute)
	out.Header = map[string]string{}

	buf := bufio.NewReader(bytes.NewBuffer(in))
	r := textproto.NewReader(buf)

	//
	// parses the method and path
	//

	line, err := r.ReadLine()
	if err != nil {
		return nil, err
	}
	parts := strings.Split(line, " ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Invalid request line. Want <method> <path>.")
	}
	out.Method = parts[0]
	out.Path = parts[1]

	//
	// parses the mime headers
	//

	header, err := r.ReadMIMEHeader()
	if err != nil {
		fmt.Println(err)
	}
	for k, v := range header {
		out.Header[k] = strings.Join(v, "; ")
	}

	//
	// extracts the response status code
	//

	out.Status, err = strconv.Atoi(header.Get("Status"))
	if err != nil {
		return nil, fmt.Errorf("Invalid Status code. %s", err)
	}
	delete(out.Header, "Status")

	//
	// parse the remainder of the file as the body
	//

	body, err := ioutil.ReadAll(buf)
	if err != nil {
		return nil, err
	}
	out.Body = string(body)

	return out, nil
}
