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

package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"go.woodpecker-ci.org/woodpecker/v3/pipeline/frontend/metadata"
)

func TestConstraintStatusSuccess(t *testing.T) {
	testdata := []struct {
		conf string
		want bool
	}{
		{conf: "", want: true},
		{conf: "{status: [failure]}", want: false},
		{conf: "{status: [success]}", want: true},
		{conf: "{status: [failure, success]}", want: true},
		{conf: "{status: {exclude: [success], include: [failure]}}", want: false},
		{conf: "{status: {exclude: [failure], include: [success]}}", want: true},
	}
	for _, test := range testdata {
		c := parseConstraints(t, test.conf)
		assert.Equal(t, test.want, c.IncludesStatusSuccess(), "when: '%s'", test.conf)
	}
}

func TestConstraints(t *testing.T) {
	testdata := []struct {
		desc string
		conf string
		with metadata.Metadata
		env  map[string]string
		want bool
	}{
		{
			desc: "no constraints, must match on default events",
			conf: "",
			with: metadata.Metadata{
				Curr: metadata.Pipeline{
					Event: metadata.EventPush,
				},
			},
			want: true,
		},
		{
			desc: "global branch filter",
			conf: "{ branch: develop }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventPush, Commit: metadata.Commit{Branch: "main"}}},
			want: false,
		},
		{
			desc: "global branch filter",
			conf: "{ branch: main }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventPush, Commit: metadata.Commit{Branch: "main"}}},
			want: true,
		},
		{
			desc: "repo constraint",
			conf: "{ repo: owner/* }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventPush}, Repo: metadata.Repo{Owner: "owner", Name: "repo"}},
			want: true,
		},
		{
			desc: "repo constraint",
			conf: "{ repo: octocat/* }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventPush}, Repo: metadata.Repo{Owner: "owner", Name: "repo"}},
			want: false,
		},
		{
			desc: "ref constraint",
			conf: "{ ref: refs/tags/* }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Commit: metadata.Commit{Ref: "refs/tags/v1.0.0"}, Event: metadata.EventPush}},
			want: true,
		},
		{
			desc: "ref constraint",
			conf: "{ ref: refs/tags/* }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Commit: metadata.Commit{Ref: "refs/heads/main"}, Event: metadata.EventPush}},
			want: false,
		},
		{
			desc: "platform constraint",
			conf: "{ platform: linux/amd64 }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventPush}, Sys: metadata.System{Platform: "linux/amd64"}},
			want: true,
		},
		{
			desc: "platform constraint",
			conf: "{ repo: linux/amd64 }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventPush}, Sys: metadata.System{Platform: "windows/amd64"}},
			want: false,
		},
		{
			desc: "instance constraint",
			conf: "{ instance: agent.tld }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventPush}, Sys: metadata.System{Host: "agent.tld"}},
			want: true,
		},
		{
			desc: "instance constraint",
			conf: "{ instance: agent.tld }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventPush}, Sys: metadata.System{Host: "beta.agent.tld"}},
			want: false,
		},
		{
			desc: "filter cron by matching name",
			conf: "{ event: cron, cron: job1 }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventCron, Cron: "job1"}},
			want: true,
		},
		{
			desc: "filter cron by name",
			conf: "{ event: cron, cron: job2 }",
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventCron, Cron: "job1"}},
			want: false,
		},
		{
			desc: "filter with build-in env passes",
			conf: "{ branch: ${CI_REPO_DEFAULT_BRANCH} }",
			with: metadata.Metadata{
				Curr: metadata.Pipeline{Event: metadata.EventPush, Commit: metadata.Commit{Branch: "stable"}},
				Repo: metadata.Repo{Branch: "stable"},
			},
			want: true,
		},
		{
			desc: "filter by eval based on event",
			conf: `{ evaluate: 'CI_PIPELINE_EVENT == "push"' }`,
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventPush}},
			want: true,
		},
		{
			desc: "filter by eval based on event and repo",
			conf: `{ evaluate: 'CI_PIPELINE_EVENT == "push" && CI_REPO == "owner/repo"' }`,
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventPush}, Repo: metadata.Repo{Owner: "owner", Name: "repo"}},
			want: true,
		},
		{
			desc: "filter by eval based on custom variable",
			conf: `{ evaluate: 'TESTVAR == "testval"' }`,
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventManual}},
			env:  map[string]string{"TESTVAR": "testval"},
			want: true,
		},
		{
			desc: "filter by eval based on custom variable",
			conf: `{ evaluate: 'TESTVAR == "testval"' }`,
			with: metadata.Metadata{Curr: metadata.Pipeline{Event: metadata.EventManual}},
			env:  map[string]string{"TESTVAR": "qwe"},
			want: false,
		},
	}

	for _, test := range testdata {
		t.Run(test.desc, func(t *testing.T) {
			conf, err := metadata.EnvVarSubst(test.conf, test.with.Environ())
			assert.NoError(t, err)
			c := parseConstraints(t, conf)
			got, err := c.Match(test.with, false, test.env)
			assert.NoError(t, err)
			assert.Equal(t, test.want, got)
		})
	}
}

func parseConstraints(t *testing.T, s string) *When {
	c := &When{}
	assert.NoError(t, yaml.Unmarshal([]byte(s), c))
	return c
}
