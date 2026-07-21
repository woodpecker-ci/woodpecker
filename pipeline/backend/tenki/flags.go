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

const (
	defaultCreateTimeout = 3 * time.Minute
	defaultMaxDuration   = 1 * time.Hour
)

var Flags = []cli.Flag{
	&cli.StringFlag{
		Name:    "backend-tenki-api-key",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_TENKI_API_KEY", "TENKI_API_KEY", "TENKI_AUTH_TOKEN"),
		Usage:   "Tenki sandbox API key",
	},
	&cli.StringFlag{
		Name:        "backend-tenki-endpoint",
		Sources:     cli.EnvVars("WOODPECKER_BACKEND_TENKI_ENDPOINT"),
		Usage:       "Tenki API base URL, leave empty for the default (production)",
		DefaultText: "Tenki production endpoint",
	},
	&cli.StringFlag{
		Name:        "backend-tenki-project-id",
		Sources:     cli.EnvVars("WOODPECKER_BACKEND_TENKI_PROJECT_ID", "TENKI_PROJECT_ID"),
		Usage:       "Tenki project to create sandboxes in; auto-resolved from the API key identity if empty",
		DefaultText: "first project of the resolved workspace",
	},
	&cli.StringFlag{
		Name:        "backend-tenki-workspace-id",
		Sources:     cli.EnvVars("WOODPECKER_BACKEND_TENKI_WORKSPACE_ID", "TENKI_WORKSPACE_ID"),
		Usage:       "Tenki workspace to scope sandboxes to; auto-resolved if empty",
		DefaultText: "first workspace of the API key identity",
	},
	&cli.BoolFlag{
		Name:    "backend-tenki-allow-outbound",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_TENKI_ALLOW_OUTBOUND"),
		Usage:   "allow outbound network access from the sandbox (needed to clone repos and fetch dependencies)",
		Value:   true,
	},
	&cli.DurationFlag{
		Name:    "backend-tenki-create-timeout",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_TENKI_CREATE_TIMEOUT"),
		Usage:   "maximum time to wait for a sandbox to become ready",
		Value:   defaultCreateTimeout,
	},
	&cli.DurationFlag{
		Name:    "backend-tenki-max-duration",
		Sources: cli.EnvVars("WOODPECKER_BACKEND_TENKI_MAX_DURATION"),
		Usage:   "maximum lifetime of a workflow sandbox before it is reclaimed",
		Value:   defaultMaxDuration,
	},
}
