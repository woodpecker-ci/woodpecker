GOFILES_NOVENDOR = $(shell find . -type f -name '*.go' -not -path "./vendor/*" -not -path "./.git/*")
GO_PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

TARGETOS ?= linux
TARGETARCH ?= amd64

VERSION ?= next
VERSION_NUMBER ?= 0.0.0
ifneq ($(CI_COMMIT_TAG),)
	VERSION := $(CI_COMMIT_TAG:v%=%)
	VERSION_NUMBER := ${VERSION}
endif

# append commit-sha to next version
BUILD_VERSION ?= $(VERSION)
ifeq ($(BUILD_VERSION),next)
	CI_COMMIT_SHA ?= $(shell git rev-parse HEAD)
	BUILD_VERSION := $(shell echo "next-$(shell echo ${CI_COMMIT_SHA} | head -c 8)")
endif

LDFLAGS := -s -w -extldflags "-static" -X github.com/woodpecker-ci/woodpecker/version.Version=${BUILD_VERSION}
CGO_ENABLED ?= 1 # only used to compile server

HAS_GO = $(shell hash go > /dev/null 2>&1 && echo "GO" || echo "NOGO" )
ifeq ($(HAS_GO),GO)
	XGO_VERSION ?= go-1.19.x
	CGO_CFLAGS ?= $(shell go env CGO_CFLAGS)
endif
CGO_CFLAGS ?=

# If the first argument is "in_docker"...
ifeq (in_docker,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "in_docker"
  MAKE_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # Ignore the next args
  $(eval $(MAKE_ARGS):;@:)

  in_docker:
	@[ "1" -eq "$(shell docker image ls woodpecker/make:local -a | wc -l)" ] && docker build -f ./docker/Dockerfile.make -t woodpecker/make:local . || echo reuse existing docker image
	@echo run in docker:
	@docker run -it \
		--user $(shell id -u):$(shell id -g) \
		-e VERSION="$(VERSION)" \
		-e BUILD_VERSION="$(BUILD_VERSION)" \
		-e CI_COMMIT_SHA="$(CI_COMMIT_SHA)" \
		-e GO_PACKAGES="$(GO_PACKAGES)" \
		-e TARGETOS="$(TARGETOS)" \
		-e TARGETARCH="$(TARGETARCH)" \
		-e CGO_ENABLED="$(CGO_ENABLED)"
		-e GOPATH=/tmp/go \
		-e HOME=/tmp/home \
		-v $(PWD):/build --rm woodpecker/make:local make $(MAKE_ARGS)
else

# Proceed with normal make

##@ General

all: help

.PHONY: version
version: ## Print the current version
	@echo ${BUILD_VERSION}

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: vendor
vendor: ## Update the vendor directory
	go mod tidy
	go mod vendor

format: install-tools ## Format source code
	@gofumpt -extra -w ${GOFILES_NOVENDOR}

.PHONY: clean
clean: ## Clean build artifacts
	go clean -i ./...
	rm -rf build
	@[ "1" != "$(shell docker image ls woodpecker/make:local -a | wc -l)" ] && docker image rm woodpecker/make:local || echo no docker image to clean

check-xgo: ## Check if xgo is installed
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install src.techknowlogick.com/xgo@latest; \
	fi

install-tools: ## Install development tools
	@hash golangci-lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi ; \
	hash lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/rs/zerolog/cmd/lint@latest; \
	fi ; \
	hash gofumpt > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install mvdan.cc/gofumpt@latest; \
	fi

ui-dependencies: ## Install UI dependencies
	(cd web/; pnpm install --frozen-lockfile)

##@ Test

.PHONY: lint
lint: install-tools ## Lint code
	@echo "Running golangci-lint"
	golangci-lint run --timeout 10m
	@echo "Running zerolog linter"
	lint github.com/woodpecker-ci/woodpecker/cmd/agent
	lint github.com/woodpecker-ci/woodpecker/cmd/cli
	lint github.com/woodpecker-ci/woodpecker/cmd/server

lint-ui: ## Lint UI code
	(cd web/; pnpm install)
	(cd web/; pnpm lesshint)
	(cd web/; pnpm lint --quiet)

test-agent: ## Test agent code
	go test -race -cover -coverprofile agent-coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/agent github.com/woodpecker-ci/woodpecker/agent/...

test-server: ## Test server code
	go test -race -cover -coverprofile server-coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/server $(shell go list github.com/woodpecker-ci/woodpecker/server/... | grep -v '/store')

test-server-forgejo: ## Test only Forgejo server code
	go test -v -race -cover -coverprofile server-coverage.out -timeout 120s $(shell go list github.com/woodpecker-ci/woodpecker/server/... | grep '/forgejo')

test-cli: ## Test cli code
	go test -race -cover -coverprofile cli-coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/cli github.com/woodpecker-ci/woodpecker/cli/...

test-server-datastore: ## Test server datastore
	go test -race -timeout 30s github.com/woodpecker-ci/woodpecker/server/store/...

test-server-datastore-coverage: ## Test server datastore with coverage report
	go test -race -cover -coverprofile datastore-coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/server/store/...

test-ui: ui-dependencies ## Test UI code
	(cd web/; pnpm run lint)
	(cd web/; pnpm run formatcheck)
	(cd web/; pnpm run typecheck)
	(cd web/; pnpm run test)

test-lib: ## Test lib code
	go test -race -cover -coverprofile coverage.out -timeout 30s $(shell go list ./... | grep -v '/cmd\|/agent\|/cli\|/server')

test: test-agent test-server test-server-datastore test-cli test-lib test-ui ## Run all tests

##@ Build

build-ui: ## Build UI
	(cd web/; pnpm install --frozen-lockfile; pnpm build)

build-server: build-ui ## Build server
	CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -o dist/woodpecker-server github.com/woodpecker-ci/woodpecker/cmd/server

build-agent: ## Build agent
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -o dist/woodpecker-agent github.com/woodpecker-ci/woodpecker/cmd/agent

build-cli: ## Build cli
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -o dist/woodpecker-cli github.com/woodpecker-ci/woodpecker/cmd/cli

build: build-agent build-server build-cli ## Build all binaries

release-frontend: build-frontend ## Build frontend

cross-compile-server: ## Cross compile the server
	$(foreach platform,$(subst ;, ,$(PLATFORMS)),\
		TARGETOS=$(firstword $(subst |, ,$(platform))) \
		TARGETARCH_XGO=$(subst arm64/v8,arm64,$(subst arm/v7,arm-7,$(word 2,$(subst |, ,$(platform))))) \
		TARGETARCH_BUILDX=$(subst arm64/v8,arm64,$(subst arm/v7,arm,$(word 2,$(subst |, ,$(platform))))) \
		make release-server-xgo || exit 1; \
	)
	tree dist

release-server-xgo: check-xgo ## Create server binaries for release using xgo
	@echo "Building for:"
	@echo "os:$(TARGETOS)"
	@echo "arch orgi:$(TARGETARCH)"
	@echo "arch (xgo):$(TARGETARCH_XGO)"
	@echo "arch (buildx):$(TARGETARCH_BUILDX)"

	CGO_CFLAGS="$(CGO_CFLAGS)" xgo -go $(XGO_VERSION) -dest ./dist/server/$(TARGETOS)-$(TARGETARCH_XGO) -tags 'netgo osusergo $(TAGS)' -ldflags '-linkmode external $(LDFLAGS)' -targets '$(TARGETOS)/$(TARGETARCH_XGO)' -out woodpecker-server -pkg cmd/server .
	mkdir -p ./dist/server/$(TARGETOS)/$(TARGETARCH_BUILDX)
	mv /build/woodpecker-server-$(TARGETOS)-$(TARGETARCH_XGO) ./dist/server/$(TARGETOS)/$(TARGETARCH_BUILDX)/woodpecker-server

release-server: ## Create server binaries for release
	# compile
	GOOS=linux GOARCH=amd64 CGO_ENABLED=${CGO_ENABLED} go build -ldflags '${LDFLAGS}' -o dist/server/linux_amd64/woodpecker-server github.com/woodpecker-ci/woodpecker/cmd/server
	# tar binary files
	tar -cvzf dist/woodpecker-server_linux_amd64.tar.gz   -C dist/server/linux_amd64 woodpecker-server

release-agent: ## Create agent binaries for release
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

release-cli: ## Create cli binaries for release
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

release-tarball: ## Create tarball for release
	mkdir -p dist/
	tar -cvzf dist/woodpecker-src-$(BUILD_VERSION).tar.gz \
		agent \
		cli \
		cmd \
		go.??? \
		LICENSE \
		Makefile \
		pipeline \
		server \
		shared \
		vendor \
		version \
		woodpecker-go \
		web/index.html \
		web/node_modules \
		web/package.json \
		web/public \
		web/src \
		web/package.json \
		web/tsconfig.* \
		web/*.ts \
		web/pnpm-lock.yaml \
		web/web.go

release-checksums: ## Create checksums for all release files
	# generate shas for tar files
	(cd dist/; sha256sum *.* > checksums.txt)

release: release-frontend release-server release-agent release-cli ## Release all binaries

bundle-prepare: ## Prepare the bundles
	go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.6.0

bundle-agent: bundle-prepare ## Create bundles for agent
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-agent.yml --target ./dist --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-agent.yml --target ./dist --packager rpm

bundle-server: bundle-prepare ## Create bundles for server
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-server.yml --target ./dist --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-server.yml --target ./dist --packager rpm

bundle-cli: bundle-prepare ## Create bundles for cli
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-cli.yml --target ./dist --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-cli.yml --target ./dist --packager rpm

bundle: bundle-agent bundle-server bundle-cli ## Create all bundles

##@ Docs
.PHONY: docs
docs: ## Generate docs (currently only for the cli)
	go generate cmd/cli/app.go

endif
