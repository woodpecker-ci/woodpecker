DOCKER_RUN_GO_VERSION=1.16
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./.git/*")
GO_PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

VERSION ?= ${DRONE_TAG}
ifeq ($(VERSION),)
	VERSION := $(shell echo ${DRONE_COMMIT_SHA} | head -c 8)
endif

LDFLAGS ?= -extldflags "-static"
ifneq ($(VERSION),)
	LDFLAGS := ${LDFLAGS} -X github.com/woodpecker-ci/woodpecker/version.Version=${VERSION}
endif


DOCKER_RUN?=
_with-docker:
	$(eval DOCKER_RUN=docker run --rm -v $(shell pwd):/go/src/ -v $(shell pwd)/build:/build -w /go/src golang:$(DOCKER_RUN_GO_VERSION))

all: build

vendor:
	go mod tidy
	go mod vendor

formatcheck:
	@([ -z "$(shell gofmt -d $(GOFILES_NOVENDOR) | head)" ]) || (echo "Source is unformatted"; exit 1)

format:
	@gofmt -w ${GOFILES_NOVENDOR}

.PHONY: clean
clean:
	go clean -i ./...
	rm -rf build

.PHONY: vet
vet:
	@echo "Running go vet..."
	@go vet $(GO_PACKAGES)

test-agent:
	$(DOCKER_RUN) go test -race -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/agent $(GO_PACKAGES)

test-server:
	$(DOCKER_RUN) go test -race -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/server

test-frontend:
	(cd web/; yarn run test)

test-lib:
	$(DOCKER_RUN) go test -race -timeout 30s $(shell go list ./... | grep -v '/cmd/')

test: test-lib test-agent test-server

build-agent:
	$(DOCKER_RUN) go build -o build/woodpecker-agent github.com/woodpecker-ci/woodpecker/cmd/agent

build-server:
	$(DOCKER_RUN) go build -o build/woodpecker-server github.com/woodpecker-ci/woodpecker/cmd/server

build-frontend:
	(cd web/; yarn run build)

build: build-agent build-server

release-agent:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/woodpecker-agent github.com/woodpecker-ci/woodpecker/cmd/agent

release-server:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags '${LDFLAGS}' -o release/woodpecker-server github.com/woodpecker-ci/woodpecker/cmd/server

release-cli:
	# disable CGO for cross-compiling
	export CGO_ENABLED=0
	# compile for all architectures
	GOOS=linux   GOARCH=amd64 go build -ldflags '${LDFLAGS}' -o cli/release/linux/amd64/woodpecker-cli   github.com/woodpecker-ci/woodpecker/cmd/cli
	GOOS=linux   GOARCH=arm64 go build -ldflags '${LDFLAGS}' -o cli/release/linux/arm64/woodpecker-cli   github.com/woodpecker-ci/woodpecker/cmd/cli
	GOOS=linux   GOARCH=arm   go build -ldflags '${LDFLAGS}' -o cli/release/linux/arm/woodpecker-cli     github.com/woodpecker-ci/woodpecker/cmd/cli
	GOOS=windows GOARCH=amd64 go build -ldflags '${LDFLAGS}' -o cli/release/windows/amd64/woodpecker-cli github.com/woodpecker-ci/woodpecker/cmd/cli
	GOOS=darwin  GOARCH=amd64 go build -ldflags '${LDFLAGS}' -o cli/release/darwin/amd64/woodpecker-cli  github.com/woodpecker-ci/woodpecker/cmd/cli
	# tar binary files prior to upload
	tar -cvzf cli/release/woodpecker_linux_amd64.tar.gz   -C cli/release/linux/amd64   woodpecker-cli
	tar -cvzf cli/release/woodpecker_linux_arm64.tar.gz   -C cli/release/linux/arm64   woodpecker-cli
	tar -cvzf cli/release/woodpecker_linux_arm.tar.gz     -C cli/release/linux/arm     woodpecker-cli
	tar -cvzf cli/release/woodpecker_windows_amd64.tar.gz -C cli/release/windows/amd64 woodpecker-cli
	tar -cvzf cli/release/woodpecker_darwin_amd64.tar.gz  -C cli/release/darwin/amd64  woodpecker-cli
	# generate shas for tar files
	sha256sum cli/release/*.tar.gz > cli/release/woodpecker_checksums.txt

release: release-agent release-server
