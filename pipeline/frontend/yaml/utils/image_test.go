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

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_trimImage(t *testing.T) {
	testdata := []struct {
		from string
		want string
	}{
		{
			from: "golang",
			want: "golang",
		},
		{
			from: "golang:latest",
			want: "golang",
		},
		{
			from: "golang:1.0.0",
			want: "golang",
		},
		{
			from: "library/golang",
			want: "golang",
		},
		{
			from: "library/golang:latest",
			want: "golang",
		},
		{
			from: "library/golang:1.0.0",
			want: "golang",
		},
		{
			from: "index.docker.io/library/golang:1.0.0",
			want: "golang",
		},
		{
			from: "docker.io/library/golang:1.0.0",
			want: "golang",
		},
		{
			from: "gcr.io/library/golang:1.0.0",
			want: "gcr.io/library/golang",
		},
		// error cases, return input unmodified
		{
			from: "foo/bar?baz:boo",
			want: "foo/bar?baz:boo",
		},
	}
	for _, test := range testdata {
		assert.Equal(t, test.want, trimImage(test.from))
	}
}

func Test_expandImage(t *testing.T) {
	testdata := []struct {
		from string
		want string
	}{
		{
			from: "golang",
			want: "docker.io/library/golang:latest",
		},
		{
			from: "golang:latest",
			want: "docker.io/library/golang:latest",
		},
		{
			from: "golang:1.0.0",
			want: "docker.io/library/golang:1.0.0",
		},
		{
			from: "library/golang",
			want: "docker.io/library/golang:latest",
		},
		{
			from: "library/golang:latest",
			want: "docker.io/library/golang:latest",
		},
		{
			from: "library/golang:1.0.0",
			want: "docker.io/library/golang:1.0.0",
		},
		{
			from: "index.docker.io/library/golang:1.0.0",
			want: "docker.io/library/golang:1.0.0",
		},
		{
			from: "gcr.io/golang",
			want: "gcr.io/golang:latest",
		},
		{
			from: "gcr.io/golang:1.0.0",
			want: "gcr.io/golang:1.0.0",
		},
		{
			from: "codeberg.org/6543/hello:latest@2c98dce11f78c2b4e40f513ca82f75035eb8cfa4957a6d8eb3f917ecaf77803",
			want: "codeberg.org/6543/hello:latest@2c98dce11f78c2b4e40f513ca82f75035eb8cfa4957a6d8eb3f917ecaf77803",
		},
		// error cases, return input unmodified
		{
			from: "foo/bar?baz:boo",
			want: "foo/bar?baz:boo",
		},
	}
	for _, test := range testdata {
		assert.Equal(t, test.want, expandImage(test.from))
	}
}

func Test_imageHasTag(t *testing.T) {
	testdata := []struct {
		from string
		want bool
	}{
		{
			from: "golang",
			want: false,
		},
		{
			from: "golang:latest",
			want: true,
		},
		{
			from: "golang:1.0.0",
			want: true,
		},
		{
			from: "library/golang",
			want: false,
		},
		{
			from: "library/golang:latest",
			want: true,
		},
		{
			from: "library/golang:1.0.0",
			want: true,
		},
		{
			from: "index.docker.io/library/golang:1.0.0",
			want: true,
		},
		{
			from: "gcr.io/golang",
			want: false,
		},
		{
			from: "gcr.io/golang:1.0.0",
			want: true,
		},
		{
			from: "codeberg.org/6543/hello:latest@2c98dce11f78c2b4e40f513ca82f75035eb8cfa4957a6d8eb3f917ecaf77803",
			want: true,
		},
	}
	for _, test := range testdata {
		assert.Equal(t, test.want, imageHasTag(test.from))
	}
}

func Test_matchImage(t *testing.T) {
	testdata := []struct {
		from, to string
		want     bool
	}{
		{
			from: "golang",
			to:   "golang",
			want: true,
		},
		{
			from: "golang:latest",
			to:   "golang",
			want: true,
		},
		{
			from: "library/golang:latest",
			to:   "golang",
			want: true,
		},
		{
			from: "index.docker.io/library/golang:1.0.0",
			to:   "golang",
			want: true,
		},
		{
			from: "golang",
			to:   "golang:latest",
			want: true,
		},
		{
			from: "library/golang:latest",
			to:   "library/golang",
			want: true,
		},
		{
			from: "gcr.io/golang",
			to:   "gcr.io/golang",
			want: true,
		},
		{
			from: "gcr.io/golang:1.0.0",
			to:   "gcr.io/golang",
			want: true,
		},
		{
			from: "gcr.io/golang:latest",
			to:   "gcr.io/golang",
			want: true,
		},
		{
			from: "gcr.io/golang",
			to:   "gcr.io/golang:latest",
			want: true,
		},
		{
			from: "golang",
			to:   "library/golang",
			want: true,
		},
		{
			from: "golang",
			to:   "gcr.io/project/golang",
			want: false,
		},
		{
			from: "golang",
			to:   "gcr.io/library/golang",
			want: false,
		},
		{
			from: "golang",
			to:   "gcr.io/golang",
			want: false,
		},
		{
			from: "woodpeckerci/plugin-kaniko",
			to:   "docker.io/woodpeckerci/plugin-kaniko",
			want: true,
		},
	}
	for _, test := range testdata {
		assert.Equal(t, test.want, MatchImage(test.from, test.to))
	}
}

func Test_matchImageDynamic(t *testing.T) {
	testdata := []struct {
		name, from string
		to         []string
		want       bool
	}{
		{
			name: "simple compare",
			from: "golang",
			to:   []string{"golang"},
			want: true,
		},
		{
			name: "compare non-taged image whit list who tag requirement",
			from: "golang",
			to:   []string{"golang:v3.0"},
			want: false,
		},
		{
			name: "compare taged image whit list who tag no requirement",
			from: "golang:v3.0",
			to:   []string{"golang"},
			want: true,
		},
		{
			name: "compare taged image whit list who has image with no tag requirement",
			from: "golang:1.0",
			to:   []string{"golang", "golang:2.0"},
			want: true,
		},
		{
			name: "compare taged image whit list who only has images with tag requirement",
			from: "golang:1.0",
			to:   []string{"golang:latest", "golang:2.0"},
			want: false,
		},
		{
			name: "compare taged image whit list who only has images with tag requirement",
			from: "golang:1.0",
			to:   []string{"golang:latest", "golang:1.0"},
			want: true,
		},
	}
	for _, test := range testdata {
		if !assert.Equal(t, test.want, MatchImageDynamic(test.from, test.to...)) {
			t.Logf("test data: '%s' -> '%s'", test.from, test.to)
		}
	}
}

func Test_matchHostname(t *testing.T) {
	testdata := []struct {
		image, hostname string
		want            bool
	}{
		{
			image:    "golang",
			hostname: "docker.io",
			want:     true,
		},
		{
			image:    "golang:latest",
			hostname: "docker.io",
			want:     true,
		},
		{
			image:    "golang:latest",
			hostname: "index.docker.io",
			want:     true,
		},
		{
			image:    "library/golang:latest",
			hostname: "docker.io",
			want:     true,
		},
		{
			image:    "docker.io/library/golang:1.0.0",
			hostname: "docker.io",
			want:     true,
		},
		{
			image:    "gcr.io/golang",
			hostname: "docker.io",
			want:     false,
		},
		{
			image:    "gcr.io/golang:1.0.0",
			hostname: "gcr.io",
			want:     true,
		},
		{
			image:    "1.2.3.4:8000/golang:1.0.0",
			hostname: "1.2.3.4:8000",
			want:     true,
		},
		{
			image:    "*&^%",
			hostname: "1.2.3.4:8000",
			want:     false,
		},
	}
	for _, test := range testdata {
		assert.Equal(t, test.want, MatchHostname(test.image, test.hostname))
	}
}
