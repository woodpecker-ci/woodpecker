// Copyright 2018 Drone.IO Inc.
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

// +build manual

package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
	"github.com/laszlocph/drone-oss-08/version"
	"github.com/urfave/cli"
)

func Test_main(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	app := cli.NewApp()
	app.Name = "drone-server"
	app.Version = version.Version.String()
	app.Usage = "drone server"
	app.Action = server

	flags = append(flags, cli.StringFlag{
		EnvVar: "TEST_RUN",
		Name:   "test.run",
		Usage:  "VSCode sets this flag",
	})

	app.Flags = flags
	app.Before = before

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
