package update

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/version"
)

func CheckForUpdate(ctx context.Context, force bool) (*NewVersion, error) {
	log.Debug().Msgf("Current version: %s", version.String())

	if version.String() == "dev" && !force {
		log.Debug().Msgf("Skipping update check for development version")
		return nil, nil
	}

	req, err := http.NewRequestWithContext(ctx, "GET", githubReleaseURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to fetch the latest release")
	}

	var release GithubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	// using the latest release
	if release.TagName == version.String() && !force {
		return nil, nil
	}

	log.Debug().Msgf("Latest version: %s", release.TagName)

	assetURL := ""
	fileName := fmt.Sprintf("woodpecker-cli_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	for _, asset := range release.Assets {
		if fileName == asset.Name {
			assetURL = asset.BrowserDownloadURL
			log.Debug().Msgf("Found asset for the current OS and arch: %s", assetURL)
			break
		}
	}

	if assetURL == "" {
		return nil, errors.New("no asset found for the current OS")
	}

	return &NewVersion{
		Version:  release.TagName,
		AssetURL: assetURL,
	}, nil
}

func downloadNewVersion(ctx context.Context, downloadURL string) (string, error) {
	log.Debug().Msgf("Downloading new version from %s ...", downloadURL)

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to download the new version")
	}

	file, err := os.CreateTemp("", "woodpecker-cli-*.tar.gz")
	if err != nil {
		return "", err
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		return "", err
	}

	log.Debug().Msgf("New version downloaded to %s", file.Name())

	return file.Name(), nil
}

func extractNewVersion(tarFilePath string) (string, error) {
	log.Debug().Msgf("Extracting new version from %s ...", tarFilePath)

	tarFile, err := os.Open(tarFilePath)
	if err != nil {
		return "", err
	}

	defer tarFile.Close()

	tmpDir, err := os.MkdirTemp("", "woodpecker-cli-*")
	if err != nil {
		return "", err
	}

	err = Untar(tmpDir, tarFile)
	if err != nil {
		return "", err
	}

	err = os.Remove(tarFilePath)
	if err != nil {
		return "", err
	}

	log.Debug().Msgf("New version extracted to %s", tmpDir)

	return path.Join(tmpDir, "woodpecker-cli"), nil
}
