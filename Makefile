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
CGO_CFLAGS ?=

HAS_GO = $(shell hash go > /dev/null 2>&1 && echo "GO" || echo "NOGO" )
ifeq ($(HAS_GO), GO)
	XGO_VERSION ?= go-1.18.x
	CGO_CFLAGS ?= $(shell $(GO) env CGO_CFLAGS)
endif

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
		-v $(PWD):/build --rm woodpecker/make:local make $(MAKE_ARGS)
else

# Proceed with normal make

all: build

.PHONY: vendor
vendor:
	go mod tidy
	go mod vendor

format:
	@gofmt -s -w ${GOFILES_NOVENDOR}

.PHONY: docs
docs:
	go generate cmd/cli/app.go

.PHONY: clean
clean:
	go clean -i ./...
	rm -rf build
	@[ "1" != "$(shell docker image ls woodpecker/make:local -a | wc -l)" ] && docker image rm woodpecker/make:local || echo no docker image to clean

.PHONY: lint
lint: install-tools
	@echo "Running golangci-lint"
	golangci-lint run --timeout 5m
	@echo "Running zerolog linter"
	lint github.com/woodpecker-ci/woodpecker/cmd/agent
	lint github.com/woodpecker-ci/woodpecker/cmd/cli
	lint github.com/woodpecker-ci/woodpecker/cmd/server

frontend-dependencies:
	(cd web/; yarn install --frozen-lockfile)

lint-frontend:
	(cd web/; yarn)
	(cd web/; yarn lesshint)
	(cd web/; yarn lint --quiet)

test-agent:
	go test -race -cover -coverprofile agent-coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/agent github.com/woodpecker-ci/woodpecker/agent/...

test-server:
	go test -race -cover -coverprofile server-coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/server $(shell go list github.com/woodpecker-ci/woodpecker/server/... | grep -v '/store')

test-cli:
	go test -race -cover -coverprofile cli-coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/cli github.com/woodpecker-ci/woodpecker/cli/...

test-server-datastore:
	go test -race -timeout 30s github.com/woodpecker-ci/woodpecker/server/store/...

test-server-datastore-coverage:
	go test -race -cover -coverprofile datastore-coverage.out -timeout 30s github.com/woodpecker-ci/woodpecker/server/store/...

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

install-tools:
	@hash golangci-lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi ; \
	hash lint > /dev/null 2>&1; if [ $$? -ne 0 ]; then \
		go install github.com/rs/zerolog/cmd/lint@latest; \
	fi

cross-compile-server:
	$(foreach platform,$(subst ;, ,$(PLATFORMS)),\
		TARGETOS=$(firstword $(subst |, ,$(platform))) \
		TARGETARCH_XGO=$(subst arm64/v8,arm64,$(subst arm/v7,arm-7,$(word 2,$(subst |, ,$(platform))))) \
		TARGETARCH_BUILDX=$(subst arm64/v8,arm64,$(subst arm/v7,arm,$(word 2,$(subst |, ,$(platform))))) \
		make release-server-xgo || exit 1; \
	)
	tree dist

release-server-xgo: check-xgo
	@echo "Building for:"
	@echo "os:$(TARGETOS)"
	@echo "arch orgi:$(TARGETARCH)"
	@echo "arch (xgo):$(TARGETARCH_XGO)"
	@echo "arch (buildx):$(TARGETARCH_BUILDX)"

	CGO_CFLAGS="$(CGO_CFLAGS)" xgo -go $(XGO_VERSION) -dest ./dist/server/$(TARGETOS)-$(TARGETARCH_XGO) -tags 'netgo osusergo $(TAGS)' -ldflags '-linkmode external $(LDFLAGS)' -targets '$(TARGETOS)/$(TARGETARCH_XGO)' -out woodpecker-server -pkg cmd/server .
	mkdir -p ./dist/server/$(TARGETOS)/$(TARGETARCH_BUILDX)
	mv /build/woodpecker-server-$(TARGETOS)-$(TARGETARCH_XGO) ./dist/server/$(TARGETOS)/$(TARGETARCH_BUILDX)/woodpecker-server

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

release-tarball:
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
		web/yarn.lock \
		web/web.go

release-checksums:
	# generate shas for tar files
	(cd dist/; sha256sum *.* > checksums.txt)

release: release-frontend release-server release-agent release-cli

bundle-prepare:
	go install github.com/goreleaser/nfpm/v2/cmd/nfpm@v2.6.0

bundle-agent: bundle-prepare
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-agent.yml --target ./dist --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-agent.yml --target ./dist --packager rpm

bundle-server: bundle-prepare
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-server.yml --target ./dist --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-server.yml --target ./dist --packager rpm

bundle-cli: bundle-prepare
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-cli.yml --target ./dist --packager deb
	VERSION_NUMBER=$(VERSION_NUMBER) nfpm package --config ./nfpm/nfpm-cli.yml --target ./dist --packager rpm

bundle: bundle-agent bundle-server bundle-cli

.PHONY: version
version:
	@echo ${BUILD_VERSION}

endif
