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

	_ "github.com/joho/godotenv/autoload"
	"github.com/laszlocph/drone-oss-08/version"
	"github.com/urfave/cli"
)

func Test_main(t *testing.T) {
	os.Setenv("KUBECONFIG", "/home/laszlo/go/src/github.com/laszlocph/drone-oss-08/kubeconfig.yaml")
	os.Setenv("DRONE_HOST", "xxx")
	os.Setenv("DRONE_GITHUB", "true")
	// os.Setenv("DATABASE_CONFIG", "/var/lib/drone/drone.sqlite")
	os.Setenv("DATABASE_DRIVER", "sqlite3")
	os.Setenv("DRONE_DEBUG", "true")
	os.Setenv("DRONE_KUBERNETES", "true")
	os.Setenv("DRONE_KUBERNETES_NAMESPACE", "default")
	os.Setenv("DRONE_KUBERNETES_STORAGECLASS", "example-nfs")
	os.Setenv("DRONE_KUBERNETES_VOLUME_SIZE", "100Mi")

	os.Setenv("DRONE_ADMIN", "laszlocph")
	os.Setenv("DRONE_OPEN", "true")
	os.Setenv("DRONE_GITHUB_CLIENT", "xxx")
	os.Setenv("DRONE_GITHUB_SECRET", "xxx")

	app := cli.NewApp()
	app.Name = "drone-server"
	app.Version = version.Version.String()
	app.Usage = "drone server"
	app.Action = server
	app.Flags = flags
	app.Before = before

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
