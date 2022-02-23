GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./.git/*")
GO_PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

TARGETOS ?= linux
TARGETARCH ?= amd64

VERSION ?= next
ifneq ($(CI_COMMIT_TAG),)
	VERSION := $(CI_COMMIT_TAG:v%=%)
endif

# append commit-sha to next version
BUILD_VERSION := $(VERSION)
ifeq ($(BUILD_VERSION),next)
	CI_COMMIT_SHA ?= $(shell git rev-parse HEAD)
	BUILD_VERSION := $(shell echo "next-$(shell echo ${CI_COMMIT_SHA} | head -c 8)")
endif

LDFLAGS := -s -w -extldflags "-static" -X github.com/woodpecker-ci/woodpecker/version.Version=${BUILD_VERSION}
CGO_CFLAGS ?=

HAS_GO = $(shell hash go > /dev/null 2>&1 && echo "GO" || echo "NOGO" )
ifeq ($(HAS_GO), GO)
	XGO_VERSION ?= go-1.17.x
	CGO_CFLAGS ?= $(shell $(GO) env CGO_CFLAGS)
endif

all: build

vendor:
	go mod tidy
	go mod vendor

format:
	@gofmt -s -w ${GOFILES_NOVENDOR}

.PHONY: clean
clean:
	go clean -i ./...
	rm -rf build

.PHONY: lint
lint:
	@echo "Running golangci-lint"
	go run vendor/github.com/golangci/golangci-lint/cmd/golangci-lint/main.go run --timeout 5m
	@echo "Running zerolog linter"
	go run vendor/github.com/rs/zerolog/cmd/lint/lint.go github.com/woodpecker-ci/woodpecker/cmd/agent
	go run vendor/github.com/rs/zerolog/cmd/lint/lint.go github.com/woodpecker-ci/woodpecker/cmd/cli
	go run vendor/github.com/rs/zerolog/cmd/lint/lint.go github.com/woodpecker-ci/woodpecker/cmd/server

frontend-dependencies:
	(cd web/; yarn install --frozen-lockfile)

lint-frontend:
	(cd web/; yarn)
	(cd web/; yarn lesshint)
	(cd web/; yarn lint --quiet)

test-agent:
	go test -race -cover -coverprofile coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/agent github.com/woodpecker-ci/woodpecker/agent/...

test-server:
	go test -race -cover -coverprofile coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/server $(shell go list github.com/woodpecker-ci/woodpecker/server/... | grep -v '/store')

test-cli:
	go test -race -cover -coverprofile coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/cli github.com/woodpecker-ci/woodpecker/cli/...

test-server-datastore:
	go test -cover -coverprofile coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/server/store/...

test-frontend: frontend-dependencies
	(cd web/; yarn run lint)
	(cd web/; yarn run formatcheck)
	(cd web/; yarn run typecheck)
	(cd web/; yarn run test)

test-lib:
	go test -race -cover -coverprofile coverage.out -timeout 30s $(shell go list ./... | grep -v '/cmd\|/agent\|/cli\|/server')

test: test-agent test-server test-server-datastore test-cli test-lib test-frontend

build-frontend:
	(cd web/; yarn install --frozen-lockfile; yarn build)

build-server: build-frontend
	CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -o dist/woodpecker-server github.com/woodpecker-ci/woodpecker/cmd/server

build-agent:
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -o dist/woodpecker-agent github.com/woodpecker-ci/woodpecker/cmd/agent

build-cli:
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -o dist/woodpecker-cli github.com/woodpecker-ci/woodpecker/cmd/cli

build: build-agent build-server build-cli

release-frontend: build-frontend

check-xgo:
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install src.techknowlogick.com/xgo@latest; \
	fi

cross-compile-server: cross-compile-server-build-loop
	tree dist
	tree /build
	@$(foreach platform,$(subst ;, ,$(PLATFORMS)),TARGETOS=$(firstword $(subst |, ,$(platform))) TARGETARCH=$(word 2,$(subst |, ,$(platform))) make normalize-server-artifacts || exit 1;)

cross-compile-server-build-loop:
	$(foreach platform,$(subst ;, ,$(PLATFORMS)),TARGETOS=$(firstword $(subst |, ,$(platform))) TARGETARCH=$(word 2,$(subst |, ,$(platform))) make release-server-xgo || exit 1;)

normalize-server-artifacts:
	mv /build/woodpecker-server-$(TARGETOS)-$(TARGETARCH) dist/server/$(TARGETOS)/$(TARGETARCH)/woodpecker-server

release-server-xgo: check-xgo
	mkdir -p ./dist/server/$(TARGETOS)/$(TARGETARCH) ;\
	CGO_CFLAGS="$(CGO_CFLAGS)" xgo -go $(XGO_VERSION) -dest ./dist/server/$(TARGETOS)/$(TARGETARCH) -tags 'netgo osusergo $(TAGS)' -ldflags '-linkmode external $(LDFLAGS)' -targets '$(TARGETOS)/$(subst arm/v,arm-,$(TARGETARCH))' -out woodpecker-server -pkg cmd/server .

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
	@echo ${BUILD_VERSION}
