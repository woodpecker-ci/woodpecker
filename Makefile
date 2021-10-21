DOCKER_RUN_GO_VERSION=1.16
GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./.git/*")
GO_PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

VERSION ?= next
ifneq ($(DRONE_TAG),)
	VERSION := $(DRONE_TAG:v%=%)
endif

# append commit-sha to next version
BUILD_VERSION := $(VERSION)
ifeq ($(BUILD_VERSION),next)
	BUILD_VERSION := $(shell echo "next-$(shell echo ${DRONE_COMMIT_SHA} | head -c 8)")
endif

LDFLAGS := -s -w -extldflags "-static" -X github.com/woodpecker-ci/woodpecker/version.Version=${BUILD_VERSION}

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

.PHONY: lint
lint:
	@echo "Running zerolog linter"
	go run vendor/github.com/rs/zerolog/cmd/lint/lint.go github.com/woodpecker-ci/woodpecker/cmd/agent
	go run vendor/github.com/rs/zerolog/cmd/lint/lint.go github.com/woodpecker-ci/woodpecker/cmd/cli
	go run vendor/github.com/rs/zerolog/cmd/lint/lint.go github.com/woodpecker-ci/woodpecker/cmd/server

frontend-dependencies:
	(cd web/; yarn install --frozen-lockfile)

test-agent:
	$(DOCKER_RUN) go test -race -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/agent $(GO_PACKAGES)

test-server:
	$(DOCKER_RUN) go test -race -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/server

test-frontend: frontend-dependencies
	(cd web/; yarn run lint)
	(cd web/; yarn run formatcheck)
	(cd web/; yarn run typecheck)
	(cd web/; yarn run test)

test-lib:
	$(DOCKER_RUN) go test -race -timeout 30s $(shell go list ./... | grep -v '/cmd/')

test: test-lib test-agent test-server

build-frontend:
	(cd web/; yarn; yarn build)

build-server: build-frontend
	$(DOCKER_RUN) go build -o dist/woodpecker-server github.com/woodpecker-ci/woodpecker/cmd/server

build-agent:
	$(DOCKER_RUN) go build -o dist/woodpecker-agent github.com/woodpecker-ci/woodpecker/cmd/agent

build-cli:
	$(DOCKER_RUN) go build -o dist/woodpecker-cli github.com/woodpecker-ci/woodpecker/cmd/cli

build: build-agent build-server build-cli

release-frontend: build-frontend

release-server:
	# compile
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -ldflags '${LDFLAGS}' -o dist/server/linux_amd64/woodpecker-server github.com/woodpecker-ci/woodpecker/cmd/server
	# tar binary files
	tar -cvzf dist/woodpecker-server_linux_amd64.tar.gz   -C dist/server/linux_amd64 woodpecker-server

release-agent:
	# compile
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/agent/linux_amd64/woodpecker-agent   github.com/woodpecker-ci/woodpecker/cmd/agent
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/agent/linux_arm64/woodpecker-agent   github.com/woodpecker-ci/woodpecker/cmd/agent
	GOOS=linux   GOARCH=arm   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/agent/linux_arm/woodpecker-agent     github.com/woodpecker-ci/woodpecker/cmd/agent
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/agent/windows_amd64/woodpecker-agent github.com/woodpecker-ci/woodpecker/cmd/agent
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/agent/darwin_amd64/woodpecker-agent  github.com/woodpecker-ci/woodpecker/cmd/agent
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/agent/darwin_arm64/woodpecker-agent  github.com/woodpecker-ci/woodpecker/cmd/agent
	# tar binary files
	tar -cvzf dist/woodpecker-agent_linux_amd64.tar.gz   -C dist/agent/linux_amd64   woodpecker-agent
	tar -cvzf dist/woodpecker-agent_linux_arm64.tar.gz   -C dist/agent/linux_arm64   woodpecker-agent
	tar -cvzf dist/woodpecker-agent_linux_arm.tar.gz     -C dist/agent/linux_arm     woodpecker-agent
	tar -cvzf dist/woodpecker-agent_windows_amd64.tar.gz -C dist/agent/windows_amd64 woodpecker-agent
	tar -cvzf dist/woodpecker-agent_darwin_amd64.tar.gz  -C dist/agent/darwin_amd64  woodpecker-agent
	tar -cvzf dist/woodpecker-agent_darwin_arm64.tar.gz  -C dist/agent/darwin_arm64  woodpecker-agent

release-cli:
	# compile
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/linux_amd64/woodpecker-cli   github.com/woodpecker-ci/woodpecker/cmd/cli
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/linux_arm64/woodpecker-cli   github.com/woodpecker-ci/woodpecker/cmd/cli
	GOOS=linux   GOARCH=arm   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/linux_arm/woodpecker-cli     github.com/woodpecker-ci/woodpecker/cmd/cli
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/windows_amd64/woodpecker-cli github.com/woodpecker-ci/woodpecker/cmd/cli
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/darwin_amd64/woodpecker-cli  github.com/woodpecker-ci/woodpecker/cmd/cli
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/darwin_arm64/woodpecker-cli  github.com/woodpecker-ci/woodpecker/cmd/cli
	# tar binary files
	tar -cvzf dist/woodpecker-cli_linux_amd64.tar.gz   -C dist/cli/linux_amd64   woodpecker-cli
	tar -cvzf dist/woodpecker-cli_linux_arm64.tar.gz   -C dist/cli/linux_arm64   woodpecker-cli
	tar -cvzf dist/woodpecker-cli_linux_arm.tar.gz     -C dist/cli/linux_arm     woodpecker-cli
	tar -cvzf dist/woodpecker-cli_windows_amd64.tar.gz -C dist/cli/windows_amd64 woodpecker-cli
	tar -cvzf dist/woodpecker-cli_darwin_amd64.tar.gz  -C dist/cli/darwin_amd64  woodpecker-cli
	tar -cvzf dist/woodpecker-cli_darwin_arm64.tar.gz  -C dist/cli/darwin_arm64  woodpecker-cli

release-checksums:
	# generate shas for tar files
	(cd dist/; sha256sum *.{tar.gz,apk,deb,rpm} > checksums.txt)

release: release-frontend release-server release-agent release-cli

bundle-prepare:
	go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.6.0

bundle-agent: bundle-prepare
	nfpm package --config ./nfpm/nfpm-agent.yml --target ./dist --packager deb
	nfpm package --config ./nfpm/nfpm-agent.yml --target ./dist --packager rpm

bundle-server: bundle-prepare
	nfpm package --config ./nfpm/nfpm-server.yml --target ./dist --packager deb
	nfpm package --config ./nfpm/nfpm-server.yml --target ./dist --packager rpm

bundle-cli: bundle-prepare
	nfpm package --config ./nfpm/nfpm-cli.yml --target ./dist --packager deb
	nfpm package --config ./nfpm/nfpm-cli.yml --target ./dist --packager rpm

bundle: bundle-agent bundle-server bundle-cli

.PHONY: version
version:
	@echo ${VERSION}
