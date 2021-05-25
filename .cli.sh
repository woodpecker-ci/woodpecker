#!/bin/sh
set -e
set -x

# disable CGO for cross-compiling
export CGO_ENABLED=0

# compile for all architectures
GOOS=linux   GOARCH=amd64 go build -ldflags "-X main.version=${DRONE_TAG##v}" -o cli/release/linux/amd64/woodpecker   github.com/woodpecker-ci/woodpecker/cli/drone
GOOS=linux   GOARCH=arm64 go build -ldflags "-X main.version=${DRONE_TAG##v}" -o cli/release/linux/arm64/woodpecker   github.com/woodpecker-ci/woodpecker/cli/drone
GOOS=linux   GOARCH=arm   go build -ldflags "-X main.version=${DRONE_TAG##v}" -o cli/release/linux/arm/woodpecker     github.com/woodpecker-ci/woodpecker/cli/drone
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=${DRONE_TAG##v}" -o cli/release/windows/amd64/woodpecker github.com/woodpecker-ci/woodpecker/cli/drone
GOOS=darwin  GOARCH=amd64 go build -ldflags "-X main.version=${DRONE_TAG##v}" -o cli/release/darwin/amd64/woodpecker  github.com/woodpecker-ci/woodpecker/cli/drone

# tar binary files prior to upload
tar -cvzf cli/release/woodpecker_linux_amd64.tar.gz   -C cli/release/linux/amd64   woodpecker
tar -cvzf cli/release/woodpecker_linux_arm64.tar.gz   -C cli/release/linux/arm64   woodpecker
tar -cvzf cli/release/woodpecker_linux_arm.tar.gz     -C cli/release/linux/arm     woodpecker
tar -cvzf cli/release/woodpecker_windows_amd64.tar.gz -C cli/release/windows/amd64 woodpecker
tar -cvzf cli/release/woodpecker_darwin_amd64.tar.gz  -C cli/release/darwin/amd64  woodpecker

# generate shas for tar files
sha256sum cli/release/*.tar.gz > cli/release/woodpecker_checksums.txt
