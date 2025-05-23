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
	"testing"

	"github.com/stretchr/testify/assert"

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
	}}

	for _, testd := range testdatas {
		t.Run(testd.Title, func(t *testing.T) {
			conf, err := yaml.ParseString(testd.Data)
			assert.NoError(t, err)

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
			from: "{steps: { build: { image: plugins/docker, settings: { test: 'true' } } }, when: { branch: main, event: push } } }",
			want: "The formerly privileged plugin `plugins/docker` is no longer privileged by default, if required, add it to `WOODPECKER_PLUGINS_PRIVILEGED`",
		},
		{
			from: "{steps: { build: { image: golang, settings: { test: 'true' } } }, when: { branch: main, event: push }, clone: { git: { image: some-other/plugin-git:v1.1.0 } } }",
			want: "Specified clone image does not match allow list, netrc is not injected",
		},
		{
			from: "steps: { build: { image: golang, secrets: [ { source: mysql_username, target: mysql_username } ] } }",
			want: "Usage of `secrets` is deprecated, use `environment` in combination with `from_secret`",
		},
		{
			from: "steps: { build: { image: golang, secrets: [ 'mysql_username' ] } }",
			want: "Usage of `secrets` is deprecated, use `environment` in combination with `from_secret`",
		},
		{
			from: "steps: { build: { image: golang }, publish: { image: golang, depends_on: [ binary ] } }",
			want: "One or more of the specified dependencies do not exist",
		},
	}

	for _, test := range testdata {
		conf, err := yaml.ParseString(test.from)
		assert.NoError(t, err)

		lerr := linter.New().Lint([]*linter.WorkflowConfig{{
			File:      test.from,
			RawConfig: test.from,
			Workflow:  conf,
		}})
		assert.Error(t, lerr, "expected lint error for configuration", test.from)

		lerrors := errors.GetPipelineErrors(lerr)
		found := false
		for _, lerr := range lerrors {
			if lerr.Message == test.want {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected error %q, got %q", test.want, lerrors)
	}
}

func TestBadHabits(t *testing.T) {
	testdata := []struct {
		from string
		want string
	}{
		{
			from: "steps: { build: { image: golang } }",
			want: "Set an event filter for all steps or the entire workflow on all items of the `when` block",
		},
		{
			from: "when: [{branch: xyz}, {event: push}]\nsteps: { build: { image: golang } }",
			want: "Set an event filter for all steps or the entire workflow on all items of the `when` block",
		},
	}

	for _, test := range testdata {
		conf, err := yaml.ParseString(test.from)
		assert.NoError(t, err)

		lerr := linter.New().Lint([]*linter.WorkflowConfig{{
			File:      test.from,
			RawConfig: test.from,
			Workflow:  conf,
		}})
		assert.Error(t, lerr, "expected lint error for configuration", test.from)

		lerrors := errors.GetPipelineErrors(lerr)
		found := false
		for _, lerr := range lerrors {
			if lerr.Message == test.want {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected error %q, got %q", test.want, lerrors)
	}
}
