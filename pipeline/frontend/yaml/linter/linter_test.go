package linter

import (
	"testing"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend/yaml"
)

func TestLint(t *testing.T) {
	testdatas := []struct{ Title, Data string }{{Title: "map", Data: `
steps:
  build:
    image: docker
    privileged: true
    network_mode: host
    volumes:
      - /tmp:/tmp
    commands:
      - go build
      - go test
  publish:
    image: plugins/docker
    repo: foo/bar
    settings:
      foo: bar
services:
  redis:
    image: redis
`}, {Title: "list", Data: `
steps:
  - name: build
    image: docker
    privileged: true
    network_mode: host
    volumes:
      - /tmp:/tmp
    commands:
      - go build
      - go test
  - name: publish
    image: plugins/docker
    repo: foo/bar
    settings:
      foo: bar
`}, {
		Title: "merge maps", Data: `
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
			if err != nil {
				t.Fatalf("Cannot unmarshal yaml %q. Error: %s", testd, err)
			}
			if err := New(WithTrusted(true)).Lint(conf); err != nil {
				t.Errorf("Expected lint returns no errors, got %q", err)
			}
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
			want: "Invalid or missing pipeline section",
		},
		{
			from: "steps: { build: { image: '' }  }",
			want: "Invalid or missing image",
		},
		{
			from: "steps: { build: { image: golang, privileged: true }  }",
			want: "Insufficient privileges to use privileged mode",
		},
		{
			from: "steps: { build: { image: golang, shm_size: 10gb }  }",
			want: "Insufficient privileges to override shm_size",
		},
		{
			from: "steps: { build: { image: golang, dns: [ 8.8.8.8 ] }  }",
			want: "Insufficient privileges to use custom dns",
		},

		{
			from: "steps: { build: { image: golang, dns_search: [ example.com ] }  }",
			want: "Insufficient privileges to use dns_search",
		},
		{
			from: "steps: { build: { image: golang, devices: [ '/dev/tty0:/dev/tty0' ] }  }",
			want: "Insufficient privileges to use devices",
		},
		{
			from: "steps: { build: { image: golang, extra_hosts: [ 'somehost:162.242.195.82' ] }  }",
			want: "Insufficient privileges to use extra_hosts",
		},
		{
			from: "steps: { build: { image: golang, network_mode: host }  }",
			want: "Insufficient privileges to use network_mode",
		},
		{
			from: "steps: { build: { image: golang, networks: [ outside, default ] }  }",
			want: "Insufficient privileges to use networks",
		},
		{
			from: "steps: { build: { image: golang, volumes: [ '/opt/data:/var/lib/mysql' ] }  }",
			want: "Insufficient privileges to use volumes",
		},
		{
			from: "steps: { build: { image: golang, network_mode: 'container:name' }  }",
			want: "Insufficient privileges to use network_mode",
		},
		{
			from: "steps: { build: { image: golang, sysctls: [ net.core.somaxconn=1024 ] }  }",
			want: "Insufficient privileges to use sysctls",
		},
	}

	for _, test := range testdata {
		conf, err := yaml.ParseString(test.from)
		if err != nil {
			t.Fatalf("Cannot unmarshal yaml %q. Error: %s", test.from, err)
		}

		lerr := New().Lint(conf)
		if lerr == nil {
			t.Errorf("Expected lint error for configuration %q", test.from)
		} else if lerr.Error() != test.want {
			t.Errorf("Want error %q, got %q", test.want, lerr.Error())
		}
	}
}
