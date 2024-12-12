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
	"strings"

	"github.com/rs/zerolog/log"

	"go.woodpecker-ci.org/woodpecker/v2/version"
)

func CheckForUpdate(ctx context.Context, force bool) (*NewVersion, error) {
	return checkForUpdate(ctx, woodpeckerVersionURL, force)
}

func checkForUpdate(ctx context.Context, versionURL string, force bool) (*NewVersion, error) {
	log.Debug().Msgf("current version: %s", version.String())

	if (version.String() == "dev" || strings.HasPrefix(version.String(), "next-")) && !force {
		log.Debug().Msgf("skipping update check for development/next versions")
		return nil, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, versionURL, nil)
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

	var versionData VersionData
	if err := json.NewDecoder(resp.Body).Decode(&versionData); err != nil {
		return nil, err
	}

	upstreamVersion := versionData.Latest
	if strings.HasPrefix(version.String(), "next-") {
		upstreamVersion = versionData.Next
	} else if strings.HasSuffix(version.String(), "rc-") {
		upstreamVersion = versionData.RC
	}

	installedVersion := strings.TrimPrefix(version.Version, "v")
	upstreamVersion = strings.TrimPrefix(upstreamVersion, "v")

	// using the latest release
	if installedVersion == upstreamVersion && !force {
		log.Debug().Msgf("no new version available")
		return nil, nil
	}

	log.Debug().Msgf("new version available: %s", upstreamVersion)

	assetURL := fmt.Sprintf(githubBinaryURL, upstreamVersion, runtime.GOOS, runtime.GOARCH)
	return &NewVersion{
		Version:  upstreamVersion,
		AssetURL: assetURL,
	}, nil
}

func downloadNewVersion(ctx context.Context, downloadURL string) (string, error) {
	log.Debug().Msgf("downloading new version from %s ...", downloadURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
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

	log.Debug().Msgf("new version downloaded to %s", file.Name())

	return file.Name(), nil
}

func extractNewVersion(tarFilePath string) (string, error) {
	log.Debug().Msgf("extracting new version from %s ...", tarFilePath)

	tarFile, err := os.Open(tarFilePath)
	if err != nil {
		return "", err
	}

	defer tarFile.Close()

	tmpDir, err := os.MkdirTemp("", "woodpecker-cli-*")
	if err != nil {
		return "", err
	}

	err = UnTar(tmpDir, tarFile)
	if err != nil {
		return "", err
	}

	err = os.Remove(tarFilePath)
	if err != nil {
		return "", err
	}

	log.Debug().Msgf("new version extracted to %s", tmpDir)

	return path.Join(tmpDir, "woodpecker-cli"), nil
}
