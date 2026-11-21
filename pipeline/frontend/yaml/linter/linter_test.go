// Copyright 2023 Woodpecker Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package linter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/errors"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml"
	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/yaml/linter"
)

func TestLint(t *testing.T) {
	testdatas := []struct{ Title, Data string }{{
		Title: "map", Data: `
when:
  event: push

steps:
  build:
    image: docker
    volumes:
      - /tmp:/tmp
    commands:
      - go build
      - go test
  publish:
    image: woodpeckerci/plugin-kaniko
    settings:
      repo: foo/bar
      foo: bar
services:
  redis:
    image: redis
`,
	}, {
		Title: "list", Data: `
when:
  event: push

steps:
  - name: build
    image: docker
    volumes:
      - /tmp:/tmp
    commands:
      - go build
      - go test
  - name: publish
    image: woodpeckerci/plugin-kaniko
    settings:
      repo: foo/bar
      foo: bar
services:
  - name: redis
    image: redis
`,
	}, {
		Title: "merge maps", Data: `
when:
  event: push

variables:
  step_template: &base-step
    image: golang:1.19
    commands:
      - go version

steps:
  test base step:
    <<: *base-step
  test base step with latest image:
    <<: *base-step
    image: golang:latest
`,
	}, {
		Title: "explicitly privileged container",
		Data:  "{steps: { build: { image: plugins/docker, privileged: true, settings: { test: 'true' } } }, when: { branch: main, event: push } }",
	}, {
		Title: "signals with single exit code",
		Data:  "{steps: { check: { image: alpine, commands: [ 'exit 78' ], signals: { skip: 78 } } }, when: { event: push } }",
	}, {
		Title: "signals with list of exit codes",
		Data:  "{steps: { check: { image: alpine, commands: [ 'exit 78' ], signals: { skip: [ 78, 79 ] } } }, when: { event: push } }",
	}, {
		Title: "restart with max_restarts",
		Data:  "{steps: { flaky: { image: alpine, commands: [ 'exit 75' ], signals: { restart: { exit_codes: [ 75, 76 ], max_restarts: 5 } } } }, when: { event: push } }",
	}, {
		Title: "long form without options",
		Data:  "{steps: { check: { image: alpine, commands: [ 'exit 78' ], signals: { skip: { exit_codes: 78 } } } }, when: { event: push } }",
	}, {
		Title: "all signals in one step",
		Data:  "{steps: { check: { image: alpine, commands: [ 'exit 78' ], signals: { skip: 78, success: [ 1, 2 ], restart: 75, cancel: 70 } } }, when: { event: push } }",
	}}

	for _, testd := range testdatas {
		t.Run(testd.Title, func(t *testing.T) {
			conf, err := yaml.ParseString(testd.Data)
			require.NoError(t, err)

			assert.NoError(t, linter.New(linter.WithTrusted(linter.TrustedConfiguration{
				Network:  true,
				Volumes:  true,
				Security: true,
			})).Lint([]*linter.WorkflowConfig{{
				File:      testd.Title,
				RawConfig: testd.Data,
				Workflow:  conf,
			}}), "expected lint returns no errors")
		})
	}
}

func TestLintDeeplyNestedBackendOptions(t *testing.T) {
	config := `
when:
  event: push
steps:
  test:
    image: golang
    backend_options:
      kubernetes:
        affinity:
          nodeAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
              nodeSelectorTerms:
                - matchExpressions:
                    - key: accelerator
                      operator: In
                      values:
                        - nvidia-tesla-v100
`

	workflow, err := yaml.ParseString(config)
	require.NoError(t, err)

	err = linter.New().Lint([]*linter.WorkflowConfig{{
		File:      "deep.yml",
		RawConfig: config,
		Workflow:  workflow,
	}})
	require.NoError(t, err)
}

func TestLintErrors(t *testing.T) {
	testdata := []struct {
		from string
		want string
	}{
		{
			from: "",
			want: "Invalid or missing `steps` section",
		},
		{
			from: "steps: { build: { image: '' }  }",
			want: "Invalid or missing image",
		},
		{
			from: "steps: { build: { image: golang, privileged: true }  }",
			want: "Insufficient trust level to use `privileged` mode",
		},
		{
			from: "steps: { build: { image: golang, dns: [ 8.8.8.8 ] }  }",
			want: "Insufficient trust level to use custom `dns`",
		},

		{
			from: "steps: { build: { image: golang, dns_search: [ example.com ] }  }",
			want: "Insufficient trust level to use `dns_search`",
		},
		{
			from: "steps: { build: { image: golang, devices: [ '/dev/tty0:/dev/tty0' ] }  }",
			want: "Insufficient trust level to use `devices`",
		},
		{
			from: "steps: { build: { image: golang, extra_hosts: [ 'somehost:162.242.195.82' ] }  }",
			want: "Insufficient trust level to use `extra_hosts`",
		},
		{
			from: "steps: { build: { image: golang, network_mode: host }  }",
			want: "Insufficient trust level to use `network_mode`",
		},
		{
			from: "steps: { build: { image: golang, volumes: [ '/opt/data:/var/lib/mysql' ] }  }",
			want: "Insufficient trust level to use `volumes`",
		},
		{
			from: "steps: { build: { image: golang, network_mode: 'container:name' }  }",
			want: "Insufficient trust level to use `network_mode`",
		},
		{
			from: "steps: { build: { image: golang, settings: { test: 'true' }, commands: [ 'echo ja', 'echo nein' ] } }",
			want: "Cannot configure both `commands` and `settings`",
		},
		{
			from: "steps: { build: { image: golang, settings: { test: 'true' }, entrypoint: [ '/bin/fish' ] } }",
			want: "Cannot configure both `entrypoint` and `settings`",
		},
		{
			from: "steps: { build: { image: golang, settings: { test: 'true' }, environment: { 'TEST': 'true' } } }",
			want: "Should not configure both `environment` and `settings`",
		},
		{
			from: "{pipeline: { build: { image: golang, settings: { test: 'true' } } }, when: { branch: main, event: push } }",
			want: "Additional property pipeline is not allowed",
		},
		{
			from: "{steps: { build: { image: plugins/docker, settings: { test: 'true' } } }, when: { branch: main, event: push } }",
			want: "The formerly privileged plugin `plugins/docker` is no longer privileged by default, if required, add it to `WOODPECKER_PLUGINS_PRIVILEGED`",
		},
		{
			from: "{steps: { build: { image: golang, settings: { test: 'true' } } }, when: { branch: main, event: push }, clone: { git: { image: some-other/plugin-git:v1.1.0 } } }",
			want: "Specified clone image does not match allow list, netrc is not injected",
		},
		{
			from: "steps: { build: { image: golang }, publish: { image: golang, depends_on: [ binary ] } }",
			want: "One or more of the specified dependencies do not exist",
		},
		{
			from: "{steps: { build: { image: golang } }, services: [ { name: database, image: mysql }, { name: database, image: postgres } ] }",
			want: "Service names must be unique, `database` is used more than once",
		},
		{
			from: "steps: { check: { image: alpine, signals: { pause_forever: 78 } } }",
			want: "Unknown signal `pause_forever`",
		},
		{
			from: "steps: { check: { image: alpine, signals: { skip: 78, success: 78 } } }",
			want: "Exit code 78 is mapped to both signal `skip` and `success`",
		},
		{
			from: "steps: { db: { image: postgres, detach: true, signals: { restart: 75 } } }",
			want: "The `restart` signal is not applied to detached steps and services, its exit codes count as failure",
		},
		{
			from: "steps: { flaky: { image: alpine, signals: { restart: { exit_codes: 75, max_restarts: 11 } } } }",
			want: "`max_restarts` must be between 1 and 10",
		},
		{
			from: "steps: { flaky: { image: alpine, signals: { restart: { exit_codes: 75, max_restarts: 0 } } } }",
			want: "`max_restarts` must be between 1 and 10",
		},
		{
			from: "steps: { check: { image: alpine, signals: { skip: { exit_codes: 78, max_restarts: 2 } } } }",
			want: "`max_restarts` is only allowed for the `restart` signal",
		},
		{
			from: "steps: { check: { image: alpine, signals: { skip: [] } } }",
			want: "No exit code configured for signal `skip`",
		},
		{
			from: "steps: { check: { image: alpine, signals: { skip: 0 } } }",
			want: "Exit code 0 is out of range 1-255",
		},
		{
			from: "steps: { check: { image: alpine, signals: { skip: 256 } } }",
			want: "Exit code 256 is out of range 1-255",
		},
		{
			from: "steps: { check: { image: alpine, signals: { skip: [ 78, 78 ] } } }",
			want: "Exit code 78 is mapped to both signal `skip` and `skip`",
		},
		{
			from: "steps: { check: { image: alpine, signals: { skip: 137 } } }",
			want: "Exit code 137 is used by shells or for processes killed by a signal (e.g. 137 = OOM/SIGKILL), consider a code between 64 and 113",
		},
	}

	for _, tt := range testdata {
		conf, err := yaml.ParseString(tt.from)
		require.NoError(t, err)

		lerr := linter.New().Lint([]*linter.WorkflowConfig{{
			File:      tt.from,
			RawConfig: tt.from,
			Workflow:  conf,
		}})
		assert.Error(t, lerr, "expected lint error for configuration", tt.from)

		lerrors := errors.GetPipelineErrors(lerr)
		found := false
		for _, lerr := range lerrors {
			if lerr.Message == tt.want {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected error %q, got %q", tt.want, lerrors)
	}
}

func TestDeprecations(t *testing.T) {
	testdata := []struct {
		from string
		want string // empty = expect no deprecation warning
	}{
		{from: `steps: { build: { image: golang, commands: ["echo $CI_COMMIT_PRERELEASE"] } }`, want: "Usage of `CI_COMMIT_PRERELEASE` is deprecated, use `CI_PIPELINE_RELEASE_PRE`"},
		{from: `steps: { build: { image: golang, commands: ["echo $$CI_COMMIT_PRERELEASE"] } }`, want: "Usage of `CI_COMMIT_PRERELEASE` is deprecated, use `CI_PIPELINE_RELEASE_PRE`"},
		{from: `steps: { build: { image: golang, commands: ["echo ${CI_COMMIT_PRERELEASE}"] } }`, want: "Usage of `CI_COMMIT_PRERELEASE` is deprecated, use `CI_PIPELINE_RELEASE_PRE`"},
		{from: `steps: { build: { image: golang, commands: ["echo ${CI_COMMIT_AUTHOR_AVATAR}"] } }`, want: "Usage of `CI_COMMIT_AUTHOR_AVATAR` is deprecated, use `CI_PIPELINE_AVATAR`"},
		{from: `steps: { build: { image: golang, commands: ["echo $CI_PREV_COMMIT_AUTHOR_AVATAR"] } }`, want: "Usage of `CI_PREV_COMMIT_AUTHOR_AVATAR` is deprecated, use `CI_PREV_PIPELINE_AVATAR`"},
		// new names must not warn
		{from: `steps: { build: { image: golang, commands: ["echo $CI_PIPELINE_RELEASE_PRE"] } }`, want: ""},
		{from: `steps: { build: { image: golang, commands: ["echo $CI_PIPELINE_AVATAR"] } }`, want: ""},
		// must not match a longer var name
		{from: `steps: { build: { image: golang, commands: ["echo $CI_COMMIT_PRERELEASE_FOO"] } }`, want: ""},
		// CI_COMMIT_AUTHOR_AVATAR regexp must not fire on CI_PREV_COMMIT_AUTHOR_AVATAR only
		{from: `steps: { build: { image: golang, commands: ["echo $CI_PREV_COMMIT_AUTHOR_AVATAR"] } }`, want: "Usage of `CI_PREV_COMMIT_AUTHOR_AVATAR` is deprecated, use `CI_PREV_PIPELINE_AVATAR`"},
	}

	for _, tt := range testdata {
		conf, err := yaml.ParseString(tt.from)
		require.NoError(t, err)

		lerr := linter.New().Lint([]*linter.WorkflowConfig{{
			File:      tt.from,
			RawConfig: tt.from,
			Workflow:  conf,
		}})

		found := ""
		for _, e := range errors.GetPipelineErrors(lerr) {
			if e.Type == errors.PipelineErrorTypeDeprecation && !strings.Contains(e.Message, "runs_on") {
				found = e.Message
				break
			}
		}
		assert.Equal(t, tt.want, found, "config %q", tt.from)
	}
}

func TestBadHabits(t *testing.T) {
	testdata := []struct {
		from string
		want string
	}{
		{
			from: "steps: { build: { image: golang } }",
			want: "Consider adding a `when` block with an `event` filter to this step or the entire workflow",
		},
		{
			from: "when: [{branch: xyz}, {event: push}]\nsteps: { build: { image: golang } }",
			want: "Consider adding a `when` block with an `event` filter to this step or the entire workflow",
		},
		{
			from: "steps: { build: { image: golang, when: [{branch: main}] } }",
			want: "Set an event filter for all steps or the entire workflow on all items of the `when` block",
		},
	}

	for _, tt := range testdata {
		conf, err := yaml.ParseString(tt.from)
		require.NoError(t, err)

		lerr := linter.New().Lint([]*linter.WorkflowConfig{{
			File:      tt.from,
			RawConfig: tt.from,
			Workflow:  conf,
		}})
		assert.Error(t, lerr, "expected lint error for configuration", tt.from)

		lerrors := errors.GetPipelineErrors(lerr)
		found := false
		for _, lerr := range lerrors {
			if lerr.Message == tt.want {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected error %q, got %q", tt.want, lerrors)
	}
}
