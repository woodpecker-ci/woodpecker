package yaml

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
)

func TestConstraint(t *testing.T) {
	testdata := []struct {
		conf string
		with string
		want bool
	}{
		// string value
		{
			conf: "master",
			with: "develop",
			want: false,
		},
		{
			conf: "master",
			with: "master",
			want: true,
		},
		{
			conf: "feature/*",
			with: "feature/foo",
			want: true,
		},
		// slice value
		{
			conf: "[ master, feature/* ]",
			with: "develop",
			want: false,
		},
		{
			conf: "[ master, feature/* ]",
			with: "master",
			want: true,
		},
		{
			conf: "[ master, feature/* ]",
			with: "feature/foo",
			want: true,
		},
		// includes block
		{
			conf: "include: master",
			with: "develop",
			want: false,
		},
		{
			conf: "include: master",
			with: "master",
			want: true,
		},
		{
			conf: "include: feature/*",
			with: "master",
			want: false,
		},
		{
			conf: "include: feature/*",
			with: "feature/foo",
			want: true,
		},
		{
			conf: "include: [ master, feature/* ]",
			with: "develop",
			want: false,
		},
		{
			conf: "include: [ master, feature/* ]",
			with: "master",
			want: true,
		},
		{
			conf: "include: [ master, feature/* ]",
			with: "feature/foo",
			want: true,
		},
		// excludes block
		{
			conf: "exclude: master",
			with: "develop",
			want: true,
		},
		{
			conf: "exclude: master",
			with: "master",
			want: false,
		},
		{
			conf: "exclude: feature/*",
			with: "master",
			want: true,
		},
		{
			conf: "exclude: feature/*",
			with: "feature/foo",
			want: false,
		},
		{
			conf: "exclude: [ master, develop ]",
			with: "master",
			want: false,
		},
		{
			conf: "exclude: [ feature/*, bar ]",
			with: "master",
			want: true,
		},
		{
			conf: "exclude: [ feature/*, bar ]",
			with: "feature/foo",
			want: false,
		},
		// include and exclude blocks
		{
			conf: "{ include: [ master, feature/* ], exclude: [ develop ] }",
			with: "master",
			want: true,
		},
		{
			conf: "{ include: [ master, feature/* ], exclude: [ feature/bar ] }",
			with: "feature/bar",
			want: false,
		},
		{
			conf: "{ include: [ master, feature/* ], exclude: [ master, develop ] }",
			with: "master",
			want: false,
		},
		// empty blocks
		{
			conf: "",
			with: "master",
			want: true,
		},
	}
	for _, test := range testdata {
		c := parseConstraint(test.conf)
		got, want := c.Match(test.with), test.want
		if got != want {
			t.Errorf("Expect %q matches %q is %v", test.with, test.conf, want)
		}
	}
}

func TestConstraintList(t *testing.T) {
	testdata := []struct {
		conf    string
		with    []string
		message string
		want    bool
	}{
		{
			conf: "",
			with: []string{"CHANGELOG.md", "README.md"},
			want: true,
		},
		{
			conf: "CHANGELOG.md",
			with: []string{"CHANGELOG.md", "README.md"},
			want: true,
		},
		{
			conf: "'*.md'",
			with: []string{"CHANGELOG.md", "README.md"},
			want: true,
		},
		{
			conf: "['*.md']",
			with: []string{"CHANGELOG.md", "README.md"},
			want: true,
		},
		{
			conf: "'docs/*'",
			with: []string{"docs/README.md"},
			want: true,
		},
		{
			conf: "'docs/*'",
			with: []string{"docs/sub/README.md"},
			want: false,
		},
		{
			conf: "'docs/**'",
			with: []string{"docs/README.md", "docs/sub/README.md", "docs/sub-sub/README.md"},
			want: true,
		},
		{
			conf: "'docs/**'",
			with: []string{"README.md"},
			want: false,
		},
		{
			conf: "{ include: [ README.md ] }",
			with: []string{"CHANGELOG.md"},
			want: false,
		},
		{
			conf: "{ exclude: [ README.md ] }",
			with: []string{"design.md"},
			want: true,
		},
		// include and exclude blocks
		{
			conf: "{ include: [ '*.md', '*.ini' ], exclude: [ CHANGELOG.md ] }",
			with: []string{"README.md"},
			want: true,
		},
		{
			conf: "{ include: [ '*.md' ], exclude: [ CHANGELOG.md ] }",
			with: []string{"CHANGELOG.md"},
			want: false,
		},
		{
			conf: "{ include: [ '*.md' ], exclude: [ CHANGELOG.md ] }",
			with: []string{"README.md", "CHANGELOG.md"},
			want: false,
		},
		// commit message ignore matches
		{
			conf:    "{ include: [ README.md ], ignore_message: '[ALL]' }",
			with:    []string{"CHANGELOG.md"},
			message: "Build them [ALL]",
			want:    true,
		},
		{
			conf:    "{ exclude: [ '*.php' ], ignore_message: '[ALL]' }",
			with:    []string{"myfile.php"},
			message: "Build them [ALL]",
			want:    true,
		},
		{
			conf:    "{ ignore_message: '[ALL]' }",
			with:    []string{},
			message: "Build them [ALL]",
			want:    true,
		},
		// empty commit
		{
			conf: "{ include: [ README.md ] }",
			with: []string{},
			want: true,
		},
	}
	for _, test := range testdata {
		c := parseConstraintPath(test.conf)
		got, want := c.Match(test.with, test.message), test.want
		if got != want {
			t.Errorf("Expect %q matches %q should be %v got %v", test.with, test.conf, want, got)
		}
	}
}

func TestConstraintMap(t *testing.T) {
	testdata := []struct {
		conf string
		with map[string]string
		want bool
	}{
		{
			conf: "GOLANG: 1.7",
			with: map[string]string{"GOLANG": "1.7"},
			want: true,
		},
		{
			conf: "GOLANG: tip",
			with: map[string]string{"GOLANG": "1.7"},
			want: false,
		},
		{
			conf: "{ GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.1", "MYSQL": "5.6"},
			want: true,
		},
		{
			conf: "{ GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.0"},
			want: false,
		},
		// TODO(bradrydzewski) eventually we should enable wildcard matching
		{
			conf: "{ GOLANG: 1.7, REDIS: 3.* }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.0"},
			want: false,
		},
		// include syntax
		{
			conf: "include: { GOLANG: 1.7 }",
			with: map[string]string{"GOLANG": "1.7"},
			want: true,
		},
		{
			conf: "include: { GOLANG: tip }",
			with: map[string]string{"GOLANG": "1.7"},
			want: false,
		},
		{
			conf: "include: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.1", "MYSQL": "5.6"},
			want: true,
		},
		{
			conf: "include: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.0"},
			want: false,
		},
		// exclude syntax
		{
			conf: "exclude: { GOLANG: 1.7 }",
			with: map[string]string{"GOLANG": "1.7"},
			want: false,
		},
		{
			conf: "exclude: { GOLANG: tip }",
			with: map[string]string{"GOLANG": "1.7"},
			want: true,
		},
		{
			conf: "exclude: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.1", "MYSQL": "5.6"},
			want: false,
		},
		{
			conf: "exclude: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.0"},
			want: true,
		},
		// exclude AND include values
		{
			conf: "{ include: { GOLANG: 1.7 }, exclude: { GOLANG: 1.7 } }",
			with: map[string]string{"GOLANG": "1.7"},
			want: false,
		},
		// blanks
		{
			conf: "",
			with: map[string]string{"GOLANG": "1.7", "REDIS": "3.0"},
			want: true,
		},
		{
			conf: "GOLANG: 1.7",
			with: map[string]string{},
			want: false,
		},
		{
			conf: "{ GOLANG: 1.7, REDIS: 3.0 }",
			with: map[string]string{},
			want: false,
		},
		{
			conf: "include: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{},
			want: false,
		},
		{
			conf: "exclude: { GOLANG: 1.7, REDIS: 3.1 }",
			with: map[string]string{},
			want: true,
		},
	}
	for _, test := range testdata {
		c := parseConstraintMap(test.conf)
		got, want := c.Match(test.with), test.want
		if got != want {
			t.Errorf("Expect %q matches %q is %v", test.with, test.conf, want)
		}
	}
}

func TestConstraints(t *testing.T) {
	testdata := []struct {
		conf string
		with frontend.Metadata
		want bool
	}{
		// no constraints, must match
		{
			conf: "",
			with: frontend.Metadata{},
			want: true,
		},
		// branch constraint
		{
			conf: "{ branch: develop }",
			with: frontend.Metadata{Curr: frontend.Build{Commit: frontend.Commit{Branch: "master"}}},
			want: false,
		},
		{
			conf: "{ branch: master }",
			with: frontend.Metadata{Curr: frontend.Build{Commit: frontend.Commit{Branch: "master"}}},
			want: true,
		},
		// environment constraint
		// {
		// 	conf: "{ branch: develop }",
		// 	with: frontend.Metadata{Curr: frontend.Build{Commit: frontend.Commit{Branch: "master"}}},
		// 	want: false,
		// },
		// {
		// 	conf: "{ branch: master }",
		// 	with: frontend.Metadata{Curr: frontend.Build{Commit: frontend.Commit{Branch: "master"}}},
		// 	want: true,
		// },
		// repo constraint
		{
			conf: "{ repo: owner/* }",
			with: frontend.Metadata{Repo: frontend.Repo{Name: "owner/repo"}},
			want: true,
		},
		{
			conf: "{ repo: octocat/* }",
			with: frontend.Metadata{Repo: frontend.Repo{Name: "owner/repo"}},
			want: false,
		},
		// ref constraint
		{
			conf: "{ ref: refs/tags/* }",
			with: frontend.Metadata{Curr: frontend.Build{Commit: frontend.Commit{Ref: "refs/tags/v1.0.0"}}},
			want: true,
		},
		{
			conf: "{ ref: refs/tags/* }",
			with: frontend.Metadata{Curr: frontend.Build{Commit: frontend.Commit{Ref: "refs/heads/master"}}},
			want: false,
		},
		// platform constraint
		{
			conf: "{ platform: linux/amd64 }",
			with: frontend.Metadata{Sys: frontend.System{Arch: "linux/amd64"}},
			want: true,
		},
		{
			conf: "{ repo: linux/amd64 }",
			with: frontend.Metadata{Sys: frontend.System{Arch: "windows/amd64"}},
			want: false,
		},
		// instance constraint
		{
			conf: "{ instance: agent.tld }",
			with: frontend.Metadata{Sys: frontend.System{Host: "agent.tld"}},
			want: true,
		},
		{
			conf: "{ instance: agent.tld }",
			with: frontend.Metadata{Sys: frontend.System{Host: "beta.agent.tld"}},
			want: false,
		},
	}
	for _, test := range testdata {
		c := parseConstraints(test.conf)
		got, want := c.Match(test.with), test.want
		if got != want {
			t.Errorf("Expect %+v matches %q is %v", test.with, test.conf, want)
		}
	}
}

func parseConstraints(s string) *Constraints {
	c := &Constraints{}
	yaml.Unmarshal([]byte(s), c)
	return c
}

func parseConstraint(s string) *Constraint {
	c := &Constraint{}
	yaml.Unmarshal([]byte(s), c)
	return c
}

func parseConstraintMap(s string) *ConstraintMap {
	c := &ConstraintMap{}
	yaml.Unmarshal([]byte(s), c)
	return c
}

func parseConstraintPath(s string) *ConstraintPath {
	c := &ConstraintPath{}
	yaml.Unmarshal([]byte(s), c)
	return c
}
