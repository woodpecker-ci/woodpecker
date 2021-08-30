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

all: deps build

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
	$(DOCKER_RUN) go test -race -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/drone-agent $(GO_PACKAGES)

test-server:
	$(DOCKER_RUN) go test -race -timeout 30s github.com/woodpecker-ci/woodpecker/cmd/drone-server

test-frontend:
	(cd web/; yarn run test)

test-lib:
	$(DOCKER_RUN) go test -race -timeout 30s $(shell go list ./... | grep -v '/cmd/')

test: test-lib test-agent test-server

build-agent:
	$(DOCKER_RUN) go build -o dist/woodpecker-agent github.com/woodpecker-ci/woodpecker/cmd/drone-agent

build-frontend:
	(cd web/; yarn run build)

build-server: build-frontend
	$(DOCKER_RUN) go build -o dist/woodpecker-server github.com/woodpecker-ci/woodpecker/cmd/drone-server

build-cli:
	$(DOCKER_RUN) go build -o dist/woodpecker-cli github.com/woodpecker-ci/woodpecker/cmd/drone-server

build: build-agent build-server build-cli

.PHONY: release
release:
	goreleaser release

install:
	go install github.com/woodpecker-ci/woodpecker/cmd/drone-agent
	go install github.com/woodpecker-ci/woodpecker/cmd/drone-server
