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

	docs "github.com/urfave/cli-docs/v3"
	"github.com/urfave/cli/v3"
)

func main() {
	app := newApp()

	// fix doc string
	// TODO: remove workaround if https://github.com/urfave/cli/issues/2210 got solved
	for _, v := range app.Commands {
		if v.Name == "exec" {
			for _, f := range v.Flags {
				if f.Names()[0] == "backend-local-temp-dir" {
					flag := f.(*cli.StringFlag)
					flag.Value = "system temporary directory"
				}
			}
		}
	}

	md, err := docs.ToMarkdown(app)
	if err != nil {
		panic(err)
	}

	fi, err := os.Create("../../docs/docs/40-cli.md")
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	if _, err := fi.WriteString("# CLI\n\n" + md); err != nil {
		panic(err)
	}
}
