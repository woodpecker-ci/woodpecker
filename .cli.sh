#!/bin/sh
set -e
set -x

# disable CGO for cross-compiling
export CGO_ENABLED=0

# compile for all architectures
GOOS=linux   GOARCH=amd64 go build -ldflags "-X main.version=${DRONE_TAG##v}" -o cli/release/linux/amd64/drone   github.com/laszlocph/drone-oss-08/cli/drone
GOOS=linux   GOARCH=arm64 go build -ldflags "-X main.version=${DRONE_TAG##v}" -o cli/release/linux/arm64/drone   github.com/laszlocph/drone-oss-08/cli/drone
GOOS=linux   GOARCH=arm   go build -ldflags "-X main.version=${DRONE_TAG##v}" -o cli/release/linux/arm/drone     github.com/laszlocph/drone-oss-08/cli/drone
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.version=${DRONE_TAG##v}" -o cli/release/windows/amd64/drone github.com/laszlocph/drone-oss-08/cli/drone
GOOS=darwin  GOARCH=amd64 go build -ldflags "-X main.version=${DRONE_TAG##v}" -o cli/release/darwin/amd64/drone  github.com/laszlocph/drone-oss-08/cli/drone

# tar binary files prior to upload
tar -cvzf cli/release/drone_linux_amd64.tar.gz   -C cli/release/linux/amd64   drone
tar -cvzf cli/release/drone_linux_arm64.tar.gz   -C cli/release/linux/arm64   drone
tar -cvzf cli/release/drone_linux_arm.tar.gz     -C cli/release/linux/arm     drone
tar -cvzf cli/release/drone_windows_amd64.tar.gz -C cli/release/windows/amd64 drone
tar -cvzf cli/release/drone_darwin_amd64.tar.gz  -C cli/release/darwin/amd64  drone

# generate shas for tar files
sha256sum cli/release/*.tar.gz > cli/release/drone_checksums.txt
