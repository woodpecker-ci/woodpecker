GO_PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

TARGETOS ?= $(shell go env GOOS)
TARGETARCH ?= $(shell go env GOARCH)

BIN_SUFFIX :=
ifeq ($(TARGETOS),windows)
	BIN_SUFFIX := .exe
endif

VERSION ?= next
VERSION_NUMBER ?= 0.0.0
CI_COMMIT_SHA ?= $(shell git rev-parse HEAD)

# it's a tagged release
ifneq ($(CI_COMMIT_TAG),)
	VERSION := $(CI_COMMIT_TAG:v%=%)
	VERSION_NUMBER := ${CI_COMMIT_TAG:v%=%}
else
	# append commit-sha to next version
	ifeq ($(VERSION),next)
		VERSION := $(shell echo "next-$(shell echo ${CI_COMMIT_SHA} | cut -c -10)")
	endif
	# append commit-sha to release branch version
	ifeq ($(shell echo ${CI_COMMIT_BRANCH} | cut -c -9),release/v)
		VERSION := $(shell echo "$(shell echo ${CI_COMMIT_BRANCH} | cut -c 10-)-$(shell echo ${CI_COMMIT_SHA} | cut -c -10)")
	endif
endif

LDFLAGS := -X go.woodpecker-ci.org/woodpecker/v2/version.Version=${VERSION}
STATIC_BUILD ?= true
ifeq ($(STATIC_BUILD),true)
	LDFLAGS := -s -w -extldflags "-static" $(LDFLAGS)
endif
CGO_ENABLED ?= 1 # only used to compile server

HAS_GO = $(shell hash go > /dev/null 2>&1 && echo "GO" || echo "NOGO" )
ifeq ($(HAS_GO),GO)
	XGO_VERSION ?= go-1.20.x
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
	@[ "1" -eq "$(shell docker image ls woodpecker/make:local -a | wc -l)" ] && docker buildx build -f ./docker/Dockerfile.make -t woodpecker/make:local --load . || echo reuse existing docker image
	@echo run in docker:
	@docker run -it \
		--user $(shell id -u):$(shell id -g) \
		-e VERSION="$(VERSION)" \
		-e CI_COMMIT_SHA="$(CI_COMMIT_SHA)" \
		-e TARGETOS="$(TARGETOS)" \
		-e TARGETARCH="$(TARGETARCH)" \
		-e CGO_ENABLED="$(CGO_ENABLED)" \
		-e GOPATH=/tmp/go \
		-e HOME=/tmp/home \
		-v $(PWD):/build --rm woodpecker/make:local make $(MAKE_ARGS)
else

# Proceed with normal make

##@ General

.PHONY: all
all: help

.PHONY: version
version: ## Print the current version
	@echo ${VERSION}

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
	@gofumpt -extra -w .

.PHONY: clean
clean: ## Clean build artifacts
	go clean -i ./...
	rm -rf build
	@[ "1" != "$(shell docker image ls woodpecker/make:local -a | wc -l)" ] && docker image rm woodpecker/make:local || echo no docker image to clean

.PHONY: clean-all
clean-all: clean ## Clean all artifacts
	rm -rf dist web/dist docs/build docs/node_modules web/node_modules
	# delete generated
	rm -rf docs/docs/40-cli.md docs/swagger.json

.PHONY: generate
generate: generate-swagger ## Run all code generations
	go generate ./...

generate-swagger: install-tools ## Run swagger code generation
	swag init -g server/api/ -g cmd/server/swagger.go --outputTypes go -output cmd/server/docs
	go generate cmd/server/swagger.go

generate-license-header: install-tools
	addlicense -c "Woodpecker Authors" -ignore "vendor/**" **/*.go

check-xgo: ## Check if xgo is installed
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install src.techknowlogick.com/xgo@latest; \
	fi

install-tools: ## Install development tools
	@hash golangci-lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.56.2 ; \
	fi ; \
	hash gofumpt > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install mvdan.cc/gofumpt@latest; \
	fi ; \
	hash swag > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi ; \
	hash addlicense > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/google/addlicense@latest; \
	fi

ui-dependencies: ## Install UI dependencies
	(cd web/; pnpm install --frozen-lockfile)

##@ Test

.PHONY: lint
lint: install-tools ## Lint code
	@echo "Running golangci-lint"
	golangci-lint run

lint-ui: ui-dependencies ## Lint UI code
	(cd web/; pnpm lint --quiet)

test-agent: ## Test agent code
	go test -race -cover -coverprofile agent-coverage.out -timeout 30s go.woodpecker-ci.org/woodpecker/v2/cmd/agent go.woodpecker-ci.org/woodpecker/v2/agent/...

test-server: ## Test server code
	go test -race -cover -coverprofile server-coverage.out -timeout 30s go.woodpecker-ci.org/woodpecker/v2/cmd/server $(shell go list go.woodpecker-ci.org/woodpecker/v2/server/... | grep -v '/store')

test-cli: ## Test cli code
	go test -race -cover -coverprofile cli-coverage.out -timeout 30s go.woodpecker-ci.org/woodpecker/v2/cmd/cli go.woodpecker-ci.org/woodpecker/v2/cli/...

test-server-datastore: ## Test server datastore
	go test -timeout 120s -run TestMigrate go.woodpecker-ci.org/woodpecker/v2/server/store/...
	go test -race -timeout 30s -skip TestMigrate go.woodpecker-ci.org/woodpecker/v2/server/store/...

test-server-datastore-coverage: ## Test server datastore with coverage report
	go test -race -cover -coverprofile datastore-coverage.out -timeout 120s go.woodpecker-ci.org/woodpecker/v2/server/store/...

test-ui: ui-dependencies ## Test UI code
	(cd web/; pnpm run lint)
	(cd web/; pnpm run format:check)
	(cd web/; pnpm run typecheck)
	(cd web/; pnpm run test)

test-lib: ## Test lib code
	go test -race -cover -coverprofile coverage.out -timeout 30s $(shell go list ./... | grep -v '/cmd\|/agent\|/cli\|/server')

.PHONY: test
test: test-agent test-server test-server-datastore test-cli test-lib ## Run all tests

##@ Build

build-ui: ## Build UI
	(cd web/; pnpm install --frozen-lockfile; pnpm build)

build-server: build-ui generate-swagger ## Build server
	CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -o dist/woodpecker-server${BIN_SUFFIX} go.woodpecker-ci.org/woodpecker/v2/cmd/server

build-agent: ## Build agent
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -o dist/woodpecker-agent${BIN_SUFFIX} go.woodpecker-ci.org/woodpecker/v2/cmd/agent

build-cli: ## Build cli
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -ldflags '${LDFLAGS}' -o dist/woodpecker-cli${BIN_SUFFIX} go.woodpecker-ci.org/woodpecker/v2/cmd/cli

build-tarball: ## Build tar archive
	mkdir -p dist && tar chzvf dist/woodpecker-src.tar.gz \
	  --exclude="*.exe" \
	  --exclude="./.pnpm-store" \
	  --exclude="node_modules" \
	  --exclude="./dist" \
	  --exclude="./data" \
	  --exclude="./build" \
	  --exclude="./.git" \
	  .

.PHONY: build
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

	CGO_CFLAGS="$(CGO_CFLAGS)" xgo -go $(XGO_VERSION) -dest ./dist/server/$(TARGETOS)_$(TARGETARCH_BUILDX) -tags 'netgo osusergo grpcnotrace $(TAGS)' -ldflags '-linkmode external $(LDFLAGS)' -targets '$(TARGETOS)/$(TARGETARCH_XGO)' -out woodpecker-server -pkg cmd/server .
	@if [ "$${XGO_IN_XGO:-0}" -eq "1" ]; then echo "inside xgo image"; \
	  mkdir -p ./dist/server/$(TARGETOS)_$(TARGETARCH_BUILDX); \
	  mv -vf /build/woodpecker-server* ./dist/server/$(TARGETOS)_$(TARGETARCH_BUILDX)/woodpecker-server$(BIN_SUFFIX); \
	else echo "outside xgo image"; \
	  [ -f "./dist/server/$(TARGETOS)_$(TARGETARCH_BUILDX)/woodpecker-server$(BIN_SUFFIX)" ] && rm -v ./dist/server/$(TARGETOS)_$(TARGETARCH_BUILDX)/woodpecker-server$(BIN_SUFFIX); \
	  mv -v ./dist/server/$(TARGETOS)_$(TARGETARCH_XGO)/woodpecker-server* ./dist/server/$(TARGETOS)_$(TARGETARCH_BUILDX)/woodpecker-server$(BIN_SUFFIX); \
	fi
	@[ "$${TARGZ:-0}" -eq "1" ] && tar -cvzf dist/woodpecker-server_$(TARGETOS)_$(TARGETARCH_BUILDX).tar.gz -C dist/server/$(TARGETOS)_$(TARGETARCH_BUILDX) woodpecker-server$(BIN_SUFFIX) || echo "skip creating 'dist/woodpecker-server_$(TARGETOS)_$(TARGETARCH_BUILDX).tar.gz'"

release-server: ## Create server binaries for release
	# compile
	GOOS=$(TARGETOS) GOARCH=$(TARGETARCH) CGO_ENABLED=${CGO_ENABLED} go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o dist/server/$(TARGETOS)_$(TARGETARCH)/woodpecker-server$(BIN_SUFFIX) go.woodpecker-ci.org/woodpecker/v2/cmd/server
	# tar binary files
	tar -cvzf dist/woodpecker-server_$(TARGETOS)_$(TARGETARCH).tar.gz -C dist/server/$(TARGETOS)_$(TARGETARCH) woodpecker-server$(BIN_SUFFIX)

release-agent: ## Create agent binaries for release
	# compile
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o dist/agent/linux_amd64/woodpecker-agent       go.woodpecker-ci.org/woodpecker/v2/cmd/agent
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o dist/agent/linux_arm64/woodpecker-agent       go.woodpecker-ci.org/woodpecker/v2/cmd/agent
	GOOS=linux   GOARCH=arm   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o dist/agent/linux_arm/woodpecker-agent         go.woodpecker-ci.org/woodpecker/v2/cmd/agent
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o dist/agent/windows_amd64/woodpecker-agent.exe go.woodpecker-ci.org/woodpecker/v2/cmd/agent
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o dist/agent/darwin_amd64/woodpecker-agent      go.woodpecker-ci.org/woodpecker/v2/cmd/agent
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o dist/agent/darwin_arm64/woodpecker-agent      go.woodpecker-ci.org/woodpecker/v2/cmd/agent
	# tar binary files
	tar -cvzf dist/woodpecker-agent_linux_amd64.tar.gz   -C dist/agent/linux_amd64   woodpecker-agent
	tar -cvzf dist/woodpecker-agent_linux_arm64.tar.gz   -C dist/agent/linux_arm64   woodpecker-agent
	tar -cvzf dist/woodpecker-agent_linux_arm.tar.gz     -C dist/agent/linux_arm     woodpecker-agent
	tar -cvzf dist/woodpecker-agent_windows_amd64.tar.gz -C dist/agent/windows_amd64 woodpecker-agent.exe
	tar -cvzf dist/woodpecker-agent_darwin_amd64.tar.gz  -C dist/agent/darwin_amd64  woodpecker-agent
	tar -cvzf dist/woodpecker-agent_darwin_arm64.tar.gz  -C dist/agent/darwin_arm64  woodpecker-agent

release-cli: ## Create cli binaries for release
	# compile
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/linux_amd64/woodpecker-cli       go.woodpecker-ci.org/woodpecker/v2/cmd/cli
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/linux_arm64/woodpecker-cli       go.woodpecker-ci.org/woodpecker/v2/cmd/cli
	GOOS=linux   GOARCH=arm   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/linux_arm/woodpecker-cli         go.woodpecker-ci.org/woodpecker/v2/cmd/cli
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/windows_amd64/woodpecker-cli.exe go.woodpecker-ci.org/woodpecker/v2/cmd/cli
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/darwin_amd64/woodpecker-cli      go.woodpecker-ci.org/woodpecker/v2/cmd/cli
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o dist/cli/darwin_arm64/woodpecker-cli      go.woodpecker-ci.org/woodpecker/v2/cmd/cli
	# tar binary files
	tar -cvzf dist/woodpecker-cli_linux_amd64.tar.gz   -C dist/cli/linux_amd64   woodpecker-cli
	tar -cvzf dist/woodpecker-cli_linux_arm64.tar.gz   -C dist/cli/linux_arm64   woodpecker-cli
	tar -cvzf dist/woodpecker-cli_linux_arm.tar.gz     -C dist/cli/linux_arm     woodpecker-cli
	tar -cvzf dist/woodpecker-cli_windows_amd64.tar.gz -C dist/cli/windows_amd64 woodpecker-cli.exe
	tar -cvzf dist/woodpecker-cli_darwin_amd64.tar.gz  -C dist/cli/darwin_amd64  woodpecker-cli
	tar -cvzf dist/woodpecker-cli_darwin_arm64.tar.gz  -C dist/cli/darwin_arm64  woodpecker-cli

release-checksums: ## Create checksums for all release files
	# generate shas for tar files
	(cd dist/; sha256sum *.* > checksums.txt)

.PHONY: release
release: release-frontend release-server release-agent release-cli ## Release all binaries

bundle-prepare: ## Prepare the bundles
	go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.6.0

bundle-agent: bundle-prepare ## Create bundles for agent
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/agent.yaml --target ./dist --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/agent.yaml --target ./dist --packager rpm

bundle-server: bundle-prepare ## Create bundles for server
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/server.yaml --target ./dist --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/server.yaml --target ./dist --packager rpm

bundle-cli: bundle-prepare ## Create bundles for cli
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/cli.yaml --target ./dist --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/cli.yaml --target ./dist --packager rpm

.PHONY: bundle
bundle: bundle-agent bundle-server bundle-cli ## Create all bundles

##@ Docs
.PHONY: docs
docs: ## Generate docs (currently only for the cli)
	go generate cmd/cli/app.go
	go generate cmd/server/swagger.go

endif
