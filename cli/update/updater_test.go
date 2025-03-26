package update

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"go.woodpecker-ci.org/woodpecker/v3/version"
)

func TestCheckForUpdate(t *testing.T) {
	version.Version = "1.0.0"
	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/version.json" {
			http.NotFound(w, r)
			return
		}

		_, _ = io.WriteString(w, `{"latest": "1.0.1", "next": "1.0.2", "rc": "1.0.3"}`)
	}
	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	newVersion, err := checkForUpdate(context.Background(), ts.URL+"/version.json", false)
	if err != nil {
		t.Fatalf("Failed to check for updates: %v", err)
	}

	if newVersion == nil || newVersion.Version != "1.0.1" {
		t.Fatalf("Expected a new version 1.0.1, got: %s", newVersion)
	}
}

func TestDownloadNewVersion(t *testing.T) {
	downloadFilePath := "/woodpecker-cli_linux_amd64.tar.gz"

	fixtureHandler := func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != downloadFilePath {
			http.NotFound(w, r)
			return
		}

		_, _ = io.WriteString(w, `blob`)
	}
	ts := httptest.NewServer(http.HandlerFunc(fixtureHandler))
	defer ts.Close()

	file, err := downloadNewVersion(context.Background(), ts.URL+downloadFilePath)
	if err != nil {
		t.Fatalf("Failed to download new version: %v", err)
	}

	if file == "" {
		t.Fatalf("Expected a file path, got: %s", file)
	}

	_ = os.Remove(file)
}
