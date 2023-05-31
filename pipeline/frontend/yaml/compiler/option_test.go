package compiler

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/woodpecker-ci/woodpecker/pipeline/frontend"
)

func TestWithWorkspace(t *testing.T) {
	compiler := New(
		WithWorkspace(
			"/pipeline",
			"src/github.com/octocat/hello-world",
		),
	)
	if compiler.base != "/pipeline" {
		t.Errorf("WithWorkspace must set the base directory")
	}
	if compiler.path != "src/github.com/octocat/hello-world" {
		t.Errorf("WithWorkspace must set the path directory")
	}
}

func TestWithEscalated(t *testing.T) {
	compiler := New(
		WithEscalated(
			"docker",
			"docker-dev",
		),
	)
	if compiler.escalated[0] != "docker" || compiler.escalated[1] != "docker-dev" {
		t.Errorf("WithEscalated must whitelist privileged images")
	}
}

func TestWithVolumes(t *testing.T) {
	compiler := New(
		WithVolumes(
			"/tmp:/tmp",
			"/foo:/foo",
		),
	)
	if compiler.volumes[0] != "/tmp:/tmp" || compiler.volumes[1] != "/foo:/foo" {
		t.Errorf("TestWithVolumes must set default volumes")
	}
}

func TestWithNetworks(t *testing.T) {
	compiler := New(
		WithNetworks(
			"overlay_1",
			"overlay_bar",
		),
	)
	if compiler.networks[0] != "overlay_1" || compiler.networks[1] != "overlay_bar" {
		t.Errorf("TestWithNetworks must set networks from parameters")
	}
}

func TestWithResourceLimit(t *testing.T) {
	compiler := New(
		WithResourceLimit(
			1,
			2,
			3,
			4,
			5,
			"0,2-5",
		),
	)
	if compiler.reslimit.MemSwapLimit != 1 {
		t.Errorf("TestWithResourceLimit must set MemSwapLimit from parameters")
	}
	if compiler.reslimit.MemLimit != 2 {
		t.Errorf("TestWithResourceLimit must set MemLimit from parameters")
	}
	if compiler.reslimit.ShmSize != 3 {
		t.Errorf("TestWithResourceLimit must set ShmSize from parameters")
	}
	if compiler.reslimit.CPUQuota != 4 {
		t.Errorf("TestWithResourceLimit must set CPUQuota from parameters")
	}
	if compiler.reslimit.CPUShares != 5 {
		t.Errorf("TestWithResourceLimit must set CPUShares from parameters")
	}
	if compiler.reslimit.CPUSet != "0,2-5" {
		t.Errorf("TestWithResourceLimit must set CPUSet from parameters")
	}
}

func TestWithPrefix(t *testing.T) {
	if New(WithPrefix("someprefix_")).prefix != "someprefix_" {
		t.Errorf("WithPrefix must set the prefix")
	}
}

func TestWithMetadata(t *testing.T) {
	metadata := frontend.Metadata{
		Repo: frontend.Repo{
			Name:     "octocat/hello-world",
			Private:  true,
			Link:     "https://github.com/octocat/hello-world",
			CloneURL: "https://github.com/octocat/hello-world.git",
		},
	}
	compiler := New(
		WithMetadata(metadata),
	)
	if !reflect.DeepEqual(compiler.metadata, metadata) {
		t.Errorf("WithMetadata must set compiler the metadata")
	}

	if compiler.env["CI_REPO_NAME"] != strings.Split(metadata.Repo.Name, "/")[1] {
		t.Errorf("WithMetadata must set CI_REPO_NAME")
	}
	if compiler.env["CI_REPO_URL"] != metadata.Repo.Link {
		t.Errorf("WithMetadata must set CI_REPO_URL")
	}
	if compiler.env["CI_REPO_CLONE_URL"] != metadata.Repo.CloneURL {
		t.Errorf("WithMetadata must set CI_REPO_CLONE_URL")
	}
}

func TestWithLocal(t *testing.T) {
	if New(WithLocal(true)).local == false {
		t.Errorf("WithLocal true must enable the local flag")
	}
	if New(WithLocal(false)).local == true {
		t.Errorf("WithLocal false must disable the local flag")
	}
}

func TestWithNetrc(t *testing.T) {
	compiler := New(
		WithNetrc(
			"octocat",
			"password",
			"github.com",
		),
	)
	if compiler.cloneEnv["CI_NETRC_USERNAME"] != "octocat" {
		t.Errorf("WithNetrc should set CI_NETRC_USERNAME")
	}
	if compiler.cloneEnv["CI_NETRC_PASSWORD"] != "password" {
		t.Errorf("WithNetrc should set CI_NETRC_PASSWORD")
	}
	if compiler.cloneEnv["CI_NETRC_MACHINE"] != "github.com" {
		t.Errorf("WithNetrc should set CI_NETRC_MACHINE")
	}
}

func TestWithProxy(t *testing.T) {
	// do not execute the test if the host machine sets http proxy
	// environment variables to avoid interference with other tests.
	if noProxy != "" || httpProxy != "" || httpsProxy != "" {
		t.SkipNow()
		return
	}

	// alter the default values
	noProxy = "example.com"
	httpProxy = "bar.com"
	httpsProxy = "baz.com"

	// reset the default values
	defer func() {
		noProxy = ""
		httpProxy = ""
		httpsProxy = ""
	}()

	testdata := map[string]string{
		"no_proxy":    noProxy,
		"NO_PROXY":    noProxy,
		"http_proxy":  httpProxy,
		"HTTP_PROXY":  httpProxy,
		"https_proxy": httpsProxy,
		"HTTPS_PROXY": httpsProxy,
	}
	compiler := New(
		WithProxy(),
	)
	for key, value := range testdata {
		if compiler.env[key] != value {
			t.Errorf("WithProxy should set %s=%s", key, value)
		}
	}
}

func TestWithEnviron(t *testing.T) {
	compiler := New(
		WithEnviron(
			map[string]string{
				"RACK_ENV": "development",
				"SHOW":     "true",
			},
		),
	)
	if compiler.env["RACK_ENV"] != "development" {
		t.Errorf("WithEnviron should set RACK_ENV")
	}
	if compiler.env["SHOW"] != "true" {
		t.Errorf("WithEnviron should set SHOW")
	}
}

func TestGetenv(t *testing.T) {
	defer func() {
		os.Unsetenv("X_TEST_FOO")
		os.Unsetenv("x_test_bar")
		os.Unsetenv("x_test_baz")
	}()
	os.Setenv("X_TEST_FOO", "foo")
	os.Setenv("x_test_bar", "bar")
	os.Setenv("x_test_baz", "")
	if getenv("x_test_foo") != "foo" {
		t.Errorf("Expect X_TEST_FOO=foo")
	}
	if getenv("X_TEST_BAR") != "bar" {
		t.Errorf("Expect x_test_bar=bar")
	}
	if getenv("x_test_baz") != "" {
		t.Errorf("Expect x_test_bar=bar is empty")
	}
}

func TestWithVolumeCacher(t *testing.T) {
	compiler := New(
		WithVolumeCacher("/cache"),
	)
	cacher, ok := compiler.cacher.(*volumeCacher)
	if !ok {
		t.Errorf("Expected volume cacher configured")
	}
	if got, want := cacher.base, "/cache"; got != want {
		t.Errorf("Expected volume cacher with base %s, got %s", want, got)
	}
}

func TestWithDefaultCloneImage(t *testing.T) {
	compiler := New(
		WithDefaultCloneImage("not-an-image"),
	)
	if compiler.defaultCloneImage != "not-an-image" {
		t.Errorf("Expected default clone image 'not-an-image' not found")
	}
}

func TestWithS3Cacher(t *testing.T) {
	compiler := New(
		WithS3Cacher("some-access-key", "some-secret-key", "some-region", "some-bucket"),
	)
	cacher, ok := compiler.cacher.(*s3Cacher)
	if !ok {
		t.Errorf("Expected s3 cacher configured")
	}
	if got, want := cacher.bucket, "some-bucket"; got != want {
		t.Errorf("Expected s3 cacher with bucket %s, got %s", want, got)
	}
	if got, want := cacher.access, "some-access-key"; got != want {
		t.Errorf("Expected s3 cacher with access key %s, got %s", want, got)
	}
	if got, want := cacher.region, "some-region"; got != want {
		t.Errorf("Expected s3 cacher with region %s, got %s", want, got)
	}
	if got, want := cacher.secret, "some-secret-key"; got != want {
		t.Errorf("Expected s3 cacher with secret key %s, got %s", want, got)
	}
}
