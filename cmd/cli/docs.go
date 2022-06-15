// Copyright 2021 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build generate
// +build generate

package main

import (
	"os"
	"reflect"
	"strings"

	"github.com/urfave/cli/v2"
)

func main() {
	app := newApp()
	for _, cmd := range app.Commands {
		fixHiddenFlags(cmd)
	}
	md, err := app.ToMarkdown()
	if err != nil {
		panic(err)
	}
	// Still a bug in our version of urfave/cli/v2
	// https://github.com/urfave/cli/pull/1311
	md = md[strings.Index(md, "#"):]

	fi, err := os.Create("../../docs/docs/40-cli.md")
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	if _, err := fi.WriteString("# CLI\n\n" + md); err != nil {
		panic(err)
	}
}

// Until https://github.com/urfave/cli/pull/1346 is merged and tagged
func fixHiddenFlags(cmd *cli.Command) {
	var flags []cli.Flag
	for _, f := range cmd.Flags {
		val := reflect.Indirect(reflect.ValueOf(f)).FieldByName("Hidden")
		if !val.IsValid() || !val.Bool() {
			flags = append(flags, f)
		}
	}
	cmd.Flags = flags
	for _, sub := range cmd.Subcommands {
		fixHiddenFlags(sub)
	}
}
