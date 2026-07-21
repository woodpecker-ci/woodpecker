// Copyright 2026 Woodpecker Authors
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

package tenki

import (
	"time"

	"github.com/urfave/cli/v3"
)

type config struct {
	apiKey        string
	endpoint      string
	projectID     string
	workspaceID   string
	allowOutbound bool
	createTimeout time.Duration
	maxDuration   time.Duration
}

func configFromCli(c *cli.Command) (config, error) {
	conf := config{
		apiKey:        c.String("backend-tenki-api-key"),
		endpoint:      c.String("backend-tenki-endpoint"),
		projectID:     c.String("backend-tenki-project-id"),
		workspaceID:   c.String("backend-tenki-workspace-id"),
		allowOutbound: c.Bool("backend-tenki-allow-outbound"),
		createTimeout: c.Duration("backend-tenki-create-timeout"),
		maxDuration:   c.Duration("backend-tenki-max-duration"),
	}

	if conf.apiKey == "" {
		return conf, ErrMissingAPIKey
	}

	return conf, nil
}
