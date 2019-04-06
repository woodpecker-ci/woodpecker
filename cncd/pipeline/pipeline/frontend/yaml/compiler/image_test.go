package compiler

import (
	"testing"
)

func TestTrimImage(t *testing.T) {
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
		got, want := trimImage(test.from), test.want
		if got != want {
			t.Errorf("Want image %q trimmed to %q, got %q", test.from, want, got)
		}
	}
}

func TestExpandImage(t *testing.T) {
	testdata := []struct {
		from string
		want string
	}{
		{
			from: "golang",
			want: "golang:latest",
		},
		{
			from: "golang:latest",
			want: "golang:latest",
		},
		{
			from: "golang:1.0.0",
			want: "golang:1.0.0",
		},
		{
			from: "library/golang",
			want: "golang:latest",
		},
		{
			from: "library/golang:latest",
			want: "golang:latest",
		},
		{
			from: "library/golang:1.0.0",
			want: "golang:1.0.0",
		},
		{
			from: "index.docker.io/library/golang:1.0.0",
			want: "golang:1.0.0",
		},
		{
			from: "gcr.io/golang",
			want: "gcr.io/golang:latest",
		},
		{
			from: "gcr.io/golang:1.0.0",
			want: "gcr.io/golang:1.0.0",
		},
		// error cases, return input unmodified
		{
			from: "foo/bar?baz:boo",
			want: "foo/bar?baz:boo",
		},
	}
	for _, test := range testdata {
		got, want := expandImage(test.from), test.want
		if got != want {
			t.Errorf("Want image %q expanded to %q, got %q", test.from, want, got)
		}
	}
}

func TestMatchImage(t *testing.T) {
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
	}
	for _, test := range testdata {
		got, want := matchImage(test.from, test.to), test.want
		if got != want {
			t.Errorf("Want image %q matching %q is %v", test.from, test.to, want)
		}
	}
}

func TestMatchHostname(t *testing.T) {
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
	}
	for _, test := range testdata {
		got, want := matchHostname(test.image, test.hostname), test.want
		if got != want {
			t.Errorf("Want image %q matching hostname %q is %v", test.image, test.hostname, want)
		}
	}
}
