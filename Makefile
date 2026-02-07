# renovate: datasource=github-releases depName=mvdan/gofumpt
GOFUMPT_VERSION := v0.9.2
# renovate: datasource=github-releases depName=golangci/golangci-lint
GOLANGCI_LINT_VERSION := v2.8.0
# renovate: datasource=docker depName=docker.io/techknowlogick/xgo
XGO_VERSION := go-1.25.x

GO_PACKAGES ?= $(shell go list ./... | grep -v /vendor/)

TARGETOS ?= $(shell go env GOOS)
TARGETARCH ?= $(shell go env GOARCH)

BIN_SUFFIX :=
ifeq ($(TARGETOS),windows)
	BIN_SUFFIX := .exe
endif

DIST_DIR ?= dist

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

TAGS ?=
LDFLAGS := -X go.woodpecker-ci.org/woodpecker/v3/version.Version=${VERSION}
STATIC_BUILD ?= true
ifeq ($(STATIC_BUILD),true)
	LDFLAGS := -s -w -extldflags "-static" $(LDFLAGS)
endif
CGO_ENABLED ?= 1 # only used to compile server

HAS_GO = $(shell hash go > /dev/null 2>&1 && echo "GO" || echo "NOGO" )
ifeq ($(HAS_GO),GO)
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
		-e TARGETOS="linux" \
		-e TARGETARCH="$(TARGETARCH)" \
		-e CGO_ENABLED="$(CGO_ENABLED)" \
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

format: install-gofumpt ## Format source code
	@gofumpt -extra -w .

.PHONY: clean
clean: ## Clean build artifacts
	go clean -i ./...
	rm -rf build
	@[ "1" != "$(shell docker image ls woodpecker/make:local -a | wc -l)" ] && docker image rm woodpecker/make:local || echo no docker image to clean

.PHONY: clean-all
clean-all: clean ## Clean all artifacts
	rm -rf ${DIST_DIR} web/dist docs/build docs/node_modules web/node_modules
	# delete generated
	rm -rf docs/docs/40-cli.md docs/openapi.json

.PHONY: generate
generate: install-mockery generate-openapi ## Run all code generations
	mockery
	CGO_ENABLED=0 go generate ./...

generate-openapi: ## Run openapi code generation and format it
	CGO_ENABLED=0 go run github.com/swaggo/swag/cmd/swag fmt --exclude rpc/proto
	CGO_ENABLED=0 go generate cmd/server/openapi.go

generate-license-header: install-addlicense
	addlicense -c "Woodpecker Authors" -ignore "vendor/**" **/*.go

check-xgo: ## Check if xgo is installed
	@hash xgo > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		$(GO) install src.techknowlogick.com/xgo@latest; \
	fi

install-golangci-lint:
	@hash golangci-lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) ; \
	fi

install-gofumpt:
	@hash gofumpt > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION); \
	fi

install-addlicense:
	@hash addlicense > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/google/addlicense@latest; \
	fi

install-mockery:
	@hash mockery > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/vektra/mockery/v3@latest; \
	fi

install-protoc-gen-go:
	@hash protoc-gen-go > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install google.golang.org/protobuf/cmd/protoc-gen-go@latest; \
	fi ; \
	hash protoc-gen-go-grpc > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest; \
	fi

.PHONY: install-tools
install-tools: install-golangci-lint install-gofumpt install-addlicense install-mockery install-protoc-gen-go ## Install development tools

ui-dependencies: ## Install UI dependencies
	(cd web/; pnpm install --frozen-lockfile)

##@ Test

.PHONY: lint
lint: install-golangci-lint ## Lint code
	@echo "Running golangci-lint"
	golangci-lint run

lint-ui: ui-dependencies ## Lint UI code
	(cd web/; pnpm lint --quiet)

test-agent: ## Test agent code
	go test -race -cover -coverprofile agent-coverage.out -timeout 60s -tags 'test $(TAGS)' go.woodpecker-ci.org/woodpecker/v3/cmd/agent go.woodpecker-ci.org/woodpecker/v3/agent/...

test-server: ## Test server code
	go test -race -cover -coverprofile server-coverage.out -timeout 60s -tags 'test $(TAGS)' go.woodpecker-ci.org/woodpecker/v3/cmd/server $(shell go list go.woodpecker-ci.org/woodpecker/v3/server/... | grep -v '/store')

test-cli: ## Test cli code
	go test -race -cover -coverprofile cli-coverage.out -timeout 60s -tags 'test $(TAGS)' go.woodpecker-ci.org/woodpecker/v3/cmd/cli go.woodpecker-ci.org/woodpecker/v3/cli/...

test-server-datastore: ## Test server datastore
	go test -timeout 300s -tags 'test $(TAGS)' -run TestMigrate go.woodpecker-ci.org/woodpecker/v3/server/store/...
	go test -race -timeout 100s -tags 'test $(TAGS)' -skip TestMigrate go.woodpecker-ci.org/woodpecker/v3/server/store/...

test-server-datastore-coverage: ## Test server datastore with coverage report
	go test -race -cover -coverprofile datastore-coverage.out -timeout 300s -tags 'test $(TAGS)' go.woodpecker-ci.org/woodpecker/v3/server/store/...

test-ui: ui-dependencies ## Test UI code
	(cd web/; pnpm run lint)
	(cd web/; pnpm run format:check)
	(cd web/; pnpm run typecheck)
	(cd web/; pnpm run test)

test-lib: ## Test lib code
	go test -race -cover -coverprofile coverage.out -timeout 60s -tags 'test $(TAGS)' $(shell go list ./... | grep -v '/cmd\|/agent\|/cli\|/server')

.PHONY: test
test: test-agent test-server test-server-datastore test-cli test-lib ## Run all tests

##@ Build

build-ui: ## Build UI
	(cd web/; pnpm install --frozen-lockfile; pnpm build)

build-server: build-ui generate-openapi ## Build server
	CGO_ENABLED=${CGO_ENABLED} GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -tags '$(TAGS)' -ldflags '${LDFLAGS}' -o ${DIST_DIR}/woodpecker-server${BIN_SUFFIX} go.woodpecker-ci.org/woodpecker/v3/cmd/server

build-agent: ## Build agent
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -tags '$(TAGS)' -ldflags '${LDFLAGS}' -o ${DIST_DIR}/woodpecker-agent${BIN_SUFFIX} go.woodpecker-ci.org/woodpecker/v3/cmd/agent

build-cli: ## Build cli
	CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -tags '$(TAGS)' -ldflags '${LDFLAGS}' -o ${DIST_DIR}/woodpecker-cli${BIN_SUFFIX} go.woodpecker-ci.org/woodpecker/v3/cmd/cli

build-tarball: ## Build tar archive
	mkdir -p ${DIST_DIR} && tar chzvf ${DIST_DIR}/woodpecker-src.tar.gz \
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

.PHONY: release-frontend
release-frontend: build-ui ## Build frontend

cross-compile-server: ## Cross compile the server
	$(foreach platform,$(subst ;, ,$(PLATFORMS)),\
		TARGETOS=$(firstword $(subst |, ,$(platform))) \
		TARGETARCH_XGO=$(subst arm64/v8,arm64,$(subst arm/v7,arm-7,$(word 2,$(subst |, ,$(platform))))) \
		TARGETARCH_BUILDX=$(subst arm64/v8,arm64,$(subst arm/v7,arm,$(word 2,$(subst |, ,$(platform))))) \
		make release-server-xgo || exit 1; \
	)
	tree ${DIST_DIR}

release-server-xgo: check-xgo ## Create server binaries for release using xgo
	@echo "Building for:"
	@echo "os:$(TARGETOS)"
	@echo "arch orgi:$(TARGETARCH)"
	@echo "arch (xgo):$(TARGETARCH_XGO)"
	@echo "arch (buildx):$(TARGETARCH_BUILDX)"
	# build via xgo
	CGO_CFLAGS="$(CGO_CFLAGS)" xgo -go $(XGO_VERSION) -dest ${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH_BUILDX) -tags 'netgo osusergo grpcnotrace $(TAGS)' -ldflags '-linkmode external $(LDFLAGS)' -targets '$(TARGETOS)/$(TARGETARCH_XGO)' -out woodpecker-server -pkg cmd/server .
	# move binary into subfolder depending on target os and arch
	@if [ "$${XGO_IN_XGO:-0}" -eq "1" ]; then \
	  echo "inside xgo image"; \
	  mkdir -p ${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH_BUILDX); \
	  mv -vf /build/woodpecker-server* ${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH_BUILDX)/woodpecker-server$(BIN_SUFFIX); \
	else \
	  echo "outside xgo image"; \
	  [ -f "${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH_BUILDX)/woodpecker-server$(BIN_SUFFIX)" ] && rm -v ${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH_BUILDX)/woodpecker-server$(BIN_SUFFIX); \
	  mv -v ${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH_XGO)/woodpecker-server* ${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH_BUILDX)/woodpecker-server$(BIN_SUFFIX); \
	fi
	# if enabled package it in an archive
	@if [ "$${ARCHIVE_IT:-0}" -eq "1" ]; then \
	  if [ "$(BIN_SUFFIX)" = ".exe" ]; then \
		  rm -f  ${DIST_DIR}/woodpecker-server_$(TARGETOS)_$(TARGETARCH_BUILDX).zip; \
	    zip -j ${DIST_DIR}/woodpecker-server_$(TARGETOS)_$(TARGETARCH_BUILDX).zip ${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH_BUILDX)/woodpecker-server.exe; \
	  else \
	    tar -cvzf ${DIST_DIR}/woodpecker-server_$(TARGETOS)_$(TARGETARCH_BUILDX).tar.gz -C ${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH_BUILDX) woodpecker-server$(BIN_SUFFIX); \
	  fi; \
	else \
	  echo "skip creating '${DIST_DIR}/woodpecker-server_$(TARGETOS)_$(TARGETARCH_BUILDX).tar.gz'"; \
	fi

release-server: ## Create server binaries for release
	# compile
	GOOS=$(TARGETOS) GOARCH=$(TARGETARCH) CGO_ENABLED=${CGO_ENABLED} go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o ${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH)/woodpecker-server$(BIN_SUFFIX) go.woodpecker-ci.org/woodpecker/v3/cmd/server
	# tar binary files
	if [ "$(BIN_SUFFIX)" == ".exe" ]; then \
	  zip -j ${DIST_DIR}/woodpecker-server_$(TARGETOS)_$(TARGETARCH).zip ${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH)/woodpecker-server.exe; \
	else \
	  tar -cvzf ${DIST_DIR}/woodpecker-server_$(TARGETOS)_$(TARGETARCH).tar.gz -C ${DIST_DIR}/server/$(TARGETOS)_$(TARGETARCH) woodpecker-server$(BIN_SUFFIX); \
	fi

release-agent: ## Create agent binaries for release
	# compile
	GOOS=linux   GOARCH=amd64   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o ${DIST_DIR}/agent/linux_amd64/woodpecker-agent       go.woodpecker-ci.org/woodpecker/v3/cmd/agent
	GOOS=linux   GOARCH=arm64   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o ${DIST_DIR}/agent/linux_arm64/woodpecker-agent       go.woodpecker-ci.org/woodpecker/v3/cmd/agent
	GOOS=linux   GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o ${DIST_DIR}/agent/linux_riscv64/woodpecker-agent     go.woodpecker-ci.org/woodpecker/v3/cmd/agent
	GOOS=linux   GOARCH=arm     CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o ${DIST_DIR}/agent/linux_arm/woodpecker-agent         go.woodpecker-ci.org/woodpecker/v3/cmd/agent
	GOOS=windows GOARCH=amd64   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o ${DIST_DIR}/agent/windows_amd64/woodpecker-agent.exe go.woodpecker-ci.org/woodpecker/v3/cmd/agent
	GOOS=darwin  GOARCH=amd64   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o ${DIST_DIR}/agent/darwin_amd64/woodpecker-agent      go.woodpecker-ci.org/woodpecker/v3/cmd/agent
	GOOS=darwin  GOARCH=arm64   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -tags 'grpcnotrace $(TAGS)' -o ${DIST_DIR}/agent/darwin_arm64/woodpecker-agent      go.woodpecker-ci.org/woodpecker/v3/cmd/agent
	# tar binary files
	tar -cvzf ${DIST_DIR}/woodpecker-agent_linux_amd64.tar.gz   -C ${DIST_DIR}/agent/linux_amd64   woodpecker-agent
	tar -cvzf ${DIST_DIR}/woodpecker-agent_linux_arm64.tar.gz   -C ${DIST_DIR}/agent/linux_arm64   woodpecker-agent
	tar -cvzf ${DIST_DIR}/woodpecker-agent_linux_riscv64.tar.gz -C ${DIST_DIR}/agent/linux_riscv64 woodpecker-agent
	tar -cvzf ${DIST_DIR}/woodpecker-agent_linux_arm.tar.gz     -C ${DIST_DIR}/agent/linux_arm     woodpecker-agent
	tar -cvzf ${DIST_DIR}/woodpecker-agent_darwin_amd64.tar.gz  -C ${DIST_DIR}/agent/darwin_amd64  woodpecker-agent
	tar -cvzf ${DIST_DIR}/woodpecker-agent_darwin_arm64.tar.gz  -C ${DIST_DIR}/agent/darwin_arm64  woodpecker-agent
	# zip binary files
	rm -f  ${DIST_DIR}/woodpecker-agent_windows_amd64.zip
	zip -j ${DIST_DIR}/woodpecker-agent_windows_amd64.zip          ${DIST_DIR}/agent/windows_amd64/woodpecker-agent.exe

release-cli: ## Create cli binaries for release
	# compile
	GOOS=linux   GOARCH=amd64   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/cli/linux_amd64/woodpecker-cli       go.woodpecker-ci.org/woodpecker/v3/cmd/cli
	GOOS=linux   GOARCH=arm64   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/cli/linux_arm64/woodpecker-cli       go.woodpecker-ci.org/woodpecker/v3/cmd/cli
	GOOS=linux   GOARCH=riscv64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/cli/linux_riscv64/woodpecker-cli     go.woodpecker-ci.org/woodpecker/v3/cmd/cli
	GOOS=linux   GOARCH=arm     CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/cli/linux_arm/woodpecker-cli         go.woodpecker-ci.org/woodpecker/v3/cmd/cli
	GOOS=windows GOARCH=amd64   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/cli/windows_amd64/woodpecker-cli.exe go.woodpecker-ci.org/woodpecker/v3/cmd/cli
	GOOS=darwin  GOARCH=amd64   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/cli/darwin_amd64/woodpecker-cli      go.woodpecker-ci.org/woodpecker/v3/cmd/cli
	GOOS=darwin  GOARCH=arm64   CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o ${DIST_DIR}/cli/darwin_arm64/woodpecker-cli      go.woodpecker-ci.org/woodpecker/v3/cmd/cli
	# tar binary files
	tar -cvzf ${DIST_DIR}/woodpecker-cli_linux_amd64.tar.gz   -C ${DIST_DIR}/cli/linux_amd64   woodpecker-cli
	tar -cvzf ${DIST_DIR}/woodpecker-cli_linux_arm64.tar.gz   -C ${DIST_DIR}/cli/linux_arm64   woodpecker-cli
	tar -cvzf ${DIST_DIR}/woodpecker-cli_linux_riscv64.tar.gz -C ${DIST_DIR}/cli/linux_riscv64 woodpecker-cli
	tar -cvzf ${DIST_DIR}/woodpecker-cli_linux_arm.tar.gz     -C ${DIST_DIR}/cli/linux_arm     woodpecker-cli
	tar -cvzf ${DIST_DIR}/woodpecker-cli_darwin_amd64.tar.gz  -C ${DIST_DIR}/cli/darwin_amd64  woodpecker-cli
	tar -cvzf ${DIST_DIR}/woodpecker-cli_darwin_arm64.tar.gz  -C ${DIST_DIR}/cli/darwin_arm64  woodpecker-cli
	# zip binary files
	rm -f  ${DIST_DIR}/woodpecker-cli_windows_amd64.zip
	zip -j ${DIST_DIR}/woodpecker-cli_windows_amd64.zip          ${DIST_DIR}/cli/windows_amd64/woodpecker-cli.exe

release-checksums: ## Create checksums for all release files
	# generate shas for tar files
	(cd ${DIST_DIR}/; sha256sum *.* > checksums.txt)

.PHONY: release
release: release-frontend release-server release-agent release-cli ## Release all binaries

bundle-prepare: ## Prepare the bundles
	go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.6.0

bundle-agent: bundle-prepare ## Create bundles for agent
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/agent.yaml --target ${DIST_DIR} --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/agent.yaml --target ${DIST_DIR} --packager rpm

bundle-server: bundle-prepare ## Create bundles for server
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/server.yaml --target ${DIST_DIR} --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/server.yaml --target ${DIST_DIR} --packager rpm

bundle-cli: bundle-prepare ## Create bundles for cli
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/cli.yaml --target ${DIST_DIR} --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/cli.yaml --target ${DIST_DIR} --packager rpm

.PHONY: bundle
bundle: bundle-agent bundle-server bundle-cli ## Create all bundles

.PHONY: spellcheck
spellcheck:
	pnpx cspell lint --no-progress --gitignore '{**,.*}/{*,.*}'
	tree --gitignore \
	  -I 012_columns_rename_procs_to_steps.go \
	  -I versioned_docs -I '*opensource.svg' | \
	  pnpx cspell lint --no-progress stdin

##@ Docs
.PHONY: docs-dependencies
docs-dependencies: ## Install docs dependencies
	(cd docs/; pnpm install --frozen-lockfile)

.PHONY: generate-docs
generate-docs: ## Generate docs (currently only for the cli)
	CGO_ENABLED=0 go generate cmd/cli/app.go
	CGO_ENABLED=0 go generate cmd/server/openapi.go

.PHONY: build-docs
build-docs: generate-docs docs-dependencies ## Build the docs
	(cd docs/; pnpm build)

##@ Man Pages
.PHONY: man-cli
man-cli: ## Generate man pages for cli
	mkdir -p dist/ && CGO_ENABLED=0 go run -tags man cmd/cli/man.go cmd/cli/app.go > dist/woodpecker-cli.man.1 && gzip -9 -f dist/woodpecker-cli.man.1

.PHONY: man-agent
man-agent: ## Generate man pages for agent
	mkdir -p dist/ && CGO_ENABLED=0 go run -tags man cmd/agent/man.go > dist/woodpecker-agent.man.1 && gzip -9 -f dist/woodpecker-agent.man.1

.PHONY: man-server
man-server: ## Generate man pages for server
	mkdir -p dist/ && CGO_ENABLED=0 go run -tags man go.woodpecker-ci.org/woodpecker/v3/cmd/server > dist/woodpecker-server.man.1 && gzip -9 -f dist/woodpecker-server.man.1

.PHONY: man
man: man-cli man-agent man-server ## Generate all man pages

##@ SBOM
.PHONY: sbom
sbom: sbom-trivy sbom-syft sbom-calc sbom-clean ## Generate SBOM (runs all SBOM subtargets)
	@echo "✓ SBOM generation complete!"
	@echo ""
	@echo "=== Files Generated ==="
	@echo "  - ${DIST_DIR}/server.spdx.json"
	@echo "  - ${DIST_DIR}/agent.spdx.json"
	@echo "  - ${DIST_DIR}/cli.spdx.json"

.PHONY: sbom-trivy
sbom-trivy: ## Generate base SBOMs with Trivy (license information)
	@mkdir -p dist
	@# Check if vendor exist
	@if [ ! -d vendor ]; then \
		echo "Error: golang vendor not found. Run 'make vendor' first."; \
		exit 1; \
	fi
	@# Check if node_modules exist
	@if [ ! -d web/node_modules ]; then \
		echo "Error: WebUI node_modules not found. Run 'make build-ui' first."; \
		exit 1; \
	fi
	@echo "=== Generating base SBOM with license information ==="
	trivy fs --scanners license --license-full --format spdx-json -o ${DIST_DIR}/base.go.spdx.json go.mod
	@echo ""
	@echo "=== Generating WebUI SBOM ==="
	trivy fs --scanners license --license-full --format spdx-json -o ${DIST_DIR}/webui.spdx.json web/

.PHONY: sbom-syft
sbom-syft: ## Generate binary-specific dependency lists with Syft
	@mkdir -p dist
	@# Check if binaries exist
	@if [ ! -f ${DIST_DIR}/server/linux_amd64/woodpecker-server ]; then \
		echo "Error: Server binary not found. Run 'make release-server' first."; \
		exit 1; \
	fi
	@if [ ! -f ${DIST_DIR}/agent/linux_amd64/woodpecker-agent ]; then \
		echo "Error: Agent binary not found. Run 'make release-agent' first."; \
		exit 1; \
	fi
	@if [ ! -f ${DIST_DIR}/cli/linux_amd64/woodpecker-cli ]; then \
		echo "Error: CLI binary not found. Run 'make release-cli' first."; \
		exit 1; \
	fi
	@echo "=== Generating binary-specific dependency lists ==="
	syft scan file:${DIST_DIR}/server/linux_amd64/woodpecker-server -o spdx-json > ${DIST_DIR}/server-deps.spdx.json
	syft scan file:${DIST_DIR}/agent/linux_amd64/woodpecker-agent -o spdx-json   > ${DIST_DIR}/agent-deps.spdx.json
	syft scan file:${DIST_DIR}/cli/linux_amd64/woodpecker-cli -o spdx-json       > ${DIST_DIR}/cli-deps.spdx.json

.PHONY: sbom-calc
sbom-calc: ## Calculate and filter final SBOMs with jq
	@mkdir -p dist
	@echo "=== Filtering base SBOM for each binary ==="
	@# Filter Server
	@jq -s '.[0] as $$base | .[1] as $$binary | ([$$binary.packages[] | .SPDXID]) as $$valid_ids | $$base | .packages |= map(select(.name as $$n | $$binary.packages[] | select(.name == $$n))) | (.packages[0].SPDXID) as $$first_pkg | .relationships |= (map(select(.spdxElementId as $$e | $$valid_ids | contains([$$e])) | select(.relatedSpdxElement as $$r | ($$r == "SPDXRef-DOCUMENT") or ($$valid_ids | contains([$$r])))) + [{spdxElementId: "SPDXRef-DOCUMENT", relationshipType: "DESCRIBES", relatedSpdxElement: $$first_pkg}]) | .name = "woodpecker-server" | .creationInfo.creators += ["Tool: syft"] | .creationInfo.created = (now | strftime("%Y-%m-%dT%H:%M:%SZ"))' \
		${DIST_DIR}/base.go.spdx.json ${DIST_DIR}/server-deps.spdx.json > ${DIST_DIR}/server-go.spdx.json
	@# Filter Agent
	@jq -s '.[0] as $$base | .[1] as $$binary | ([$$binary.packages[] | .SPDXID]) as $$valid_ids | $$base | .packages |= map(select(.name as $$n | $$binary.packages[] | select(.name == $$n))) | (.packages[0].SPDXID) as $$first_pkg | .relationships |= (map(select(.spdxElementId as $$e | $$valid_ids | contains([$$e])) | select(.relatedSpdxElement as $$r | ($$r == "SPDXRef-DOCUMENT") or ($$valid_ids | contains([$$r])))) + [{spdxElementId: "SPDXRef-DOCUMENT", relationshipType: "DESCRIBES", relatedSpdxElement: $$first_pkg}]) | .name = "woodpecker-agent" | .creationInfo.creators += ["Tool: syft"] | .creationInfo.created = (now | strftime("%Y-%m-%dT%H:%M:%SZ"))' \
		${DIST_DIR}/base.go.spdx.json ${DIST_DIR}/agent-deps.spdx.json > ${DIST_DIR}/agent.spdx.json
	@# Filter CLI
	@jq -s '.[0] as $$base | .[1] as $$binary | ([$$binary.packages[] | .SPDXID]) as $$valid_ids | $$base | .packages |= map(select(.name as $$n | $$binary.packages[] | select(.name == $$n))) | (.packages[0].SPDXID) as $$first_pkg | .relationships |= (map(select(.spdxElementId as $$e | $$valid_ids | contains([$$e])) | select(.relatedSpdxElement as $$r | ($$r == "SPDXRef-DOCUMENT") or ($$valid_ids | contains([$$r])))) + [{spdxElementId: "SPDXRef-DOCUMENT", relationshipType: "DESCRIBES", relatedSpdxElement: $$first_pkg}]) | .name = "woodpecker-cli" | .creationInfo.creators += ["Tool: syft"] | .creationInfo.created = (now | strftime("%Y-%m-%dT%H:%M:%SZ"))' \
		${DIST_DIR}/base.go.spdx.json ${DIST_DIR}/cli-deps.spdx.json > ${DIST_DIR}/cli.spdx.json
	@echo ""
	@echo "=== Combining Server + WebUI ==="
	@jq -s '([.[0].packages[], .[1].packages[]] | map(.SPDXID)) as $$valid_ids | (.[0].packages[0].SPDXID) as $$first_pkg | {SPDXID: "SPDXRef-DOCUMENT", spdxVersion: .[0].spdxVersion, creationInfo: (.[0].creationInfo | .creators += ["Tool: syft"] | .created = (now | strftime("%Y-%m-%dT%H:%M:%SZ"))), name: "woodpecker-server", dataLicense: .[0].dataLicense, documentNamespace: (.[0].documentNamespace | sub("server"; "server-combined")), packages: (.[0].packages + .[1].packages), relationships: (([.[0].relationships[], .[1].relationships[]] | map(select(.spdxElementId as $$e | $$valid_ids | contains([$$e])) | select(.relatedSpdxElement as $$r | ($$r == "SPDXRef-DOCUMENT") or ($$valid_ids | contains([$$r]))))) + [{spdxElementId: "SPDXRef-DOCUMENT", relationshipType: "DESCRIBES", relatedSpdxElement: $$first_pkg}])}' \
		${DIST_DIR}/server-go.spdx.json ${DIST_DIR}/webui.spdx.json > ${DIST_DIR}/server.spdx.json
	@echo ""
	@echo "=== Package Counts ==="
	@echo "Base (all go.mod):  $$(jq '.packages | length' ${DIST_DIR}/base.go.spdx.json) packages"
	@echo "Server (combined):  $$(jq '.packages | length' ${DIST_DIR}/server.spdx.json) packages"
	@echo "  Server (Go):      $$(jq '.packages | length' ${DIST_DIR}/server-go.spdx.json) packages"
	@echo "  Server (WebUI):   $$(jq '.packages | length' ${DIST_DIR}/webui.spdx.json) packages"
	@echo "Agent binary:       $$(jq '.packages | length' ${DIST_DIR}/agent.spdx.json) packages"
	@echo "CLI binary:         $$(jq '.packages | length' ${DIST_DIR}/cli.spdx.json) packages"
	@echo ""
	@echo "=== License Coverage ==="
	@echo "Server with licenses: $$(jq '[.packages[] | select(.licenseConcluded != null and .licenseConcluded != "NOASSERTION")] | length' ${DIST_DIR}/server.spdx.json)/$$(jq '.packages | length' ${DIST_DIR}/server.spdx.json)"
	@echo "Agent with licenses:  $$(jq '[.packages[] | select(.licenseConcluded != null and .licenseConcluded != "NOASSERTION")] | length' ${DIST_DIR}/agent.spdx.json)/$$(jq '.packages | length' ${DIST_DIR}/agent.spdx.json)"
	@echo "CLI with licenses:    $$(jq '[.packages[] | select(.licenseConcluded != null and .licenseConcluded != "NOASSERTION")] | length' ${DIST_DIR}/cli.spdx.json)/$$(jq '.packages | length' ${DIST_DIR}/cli.spdx.json)"
	@echo ""
	@echo "=== License Distribution (Server) ==="
	@jq -r '[.packages[] | .licenseConcluded // "NOASSERTION"] | group_by(.) | map({license: .[0], count: length}) | sort_by(-.count) | .[] | "  \(.count | tostring | . + " " * (6 - length))\(.license)"' ${DIST_DIR}/server.spdx.json
	@echo ""
	@echo "=== License Distribution (Agent) ==="
	@jq -r '[.packages[] | .licenseConcluded // "NOASSERTION"] | group_by(.) | map({license: .[0], count: length}) | sort_by(-.count) | .[] | "  \(.count | tostring | . + " " * (6 - length))\(.license)"' ${DIST_DIR}/agent.spdx.json
	@echo ""
	@echo "=== License Distribution (CLI) ==="
	@jq -r '[.packages[] | .licenseConcluded // "NOASSERTION"] | group_by(.) | map({license: .[0], count: length}) | sort_by(-.count) | .[] | "  \(.count | tostring | . + " " * (6 - length))\(.license)"' ${DIST_DIR}/cli.spdx.json
	@echo ""

.PHONY: sbom-clean
sbom-clean:
	@echo "=== Cleaning up intermediate files ==="
	@rm -f ${DIST_DIR}/cli-deps.spdx.json ${DIST_DIR}/agent-deps.spdx.json ${DIST_DIR}/server-deps.spdx.json ${DIST_DIR}/server-go.spdx.json ${DIST_DIR}/base.go.spdx.json ${DIST_DIR}/webui.spdx.json
	@echo ""

endif
