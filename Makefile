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

deps:
	go get -u golang.org/x/net/context
	go get -u golang.org/x/net/context/ctxhttp
	go get -u github.com/golang/protobuf/proto
	go get -u github.com/golang/protobuf/protoc-gen-go
	go get -d docker.io/go-docker
	go get -d github.com/jackspirou/syscerts

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
	go get github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipec
	go get github.com/woodpecker-ci/woodpecker/remote/mock
	go get github.com/woodpecker-ci/woodpecker/cli/drone/internal
	go get -t github.com/woodpecker-ci/woodpecker/cncd/pipeline/pipeline/frontend/yaml
	$(DOCKER_RUN) go test -race -timeout 30s $(shell go list ./... | grep -v '/cmd/')

test: test-lib test-agent test-server

build-agent:
	$(DOCKER_RUN) go build -o build/drone-agent github.com/woodpecker-ci/woodpecker/cmd/drone-agent

build-server:
	$(DOCKER_RUN) go build -o build/drone-server github.com/woodpecker-ci/woodpecker/cmd/drone-server

build-frontend:
	(cd web/; yarn run build)


build: build-agent build-server

release-agent:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/drone-agent github.com/woodpecker-ci/woodpecker/cmd/drone-agent

release-server:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags '${LDFLAGS}' -o release/drone-server github.com/woodpecker-ci/woodpecker/cmd/drone-server

release: release-agent release-server

install:
	go install github.com/woodpecker-ci/woodpecker/cmd/drone-agent
	go install github.com/woodpecker-ci/woodpecker/cmd/drone-server
