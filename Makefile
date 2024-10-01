GO_PKG_NAME := github.com/openshift-eng/openshift-tests-extension

GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_TREE_STATE := $(shell if git diff --quiet; then echo clean; else echo dirty; fi)

LDFLAGS := -X '$(GO_PKG_NAME)/pkg/version.CommitFromGit=$(GIT_COMMIT)' \
           -X '$(GO_PKG_NAME)/pkg/version.BuildDate=$(BUILD_DATE)' \
           -X '$(GO_PKG_NAME)/pkg/version.GitTreeState=$(GIT_TREE_STATE)'

.PHONY: verify test lint clean unit integration

all: unit build integration

verify: lint

build: example-tests framework-tests

example-tests:
	go build -ldflags "$(LDFLAGS)" ./cmd/example-tests/...

framework-tests:
	go build -ldflags "$(LDFLAGS)" ./cmd/framework-tests/...

test: unit integration

unit:
	go test ./...

integration: build
	./framework-tests run-suite framework

lint:
	./hack/go-lint.sh run ./...

clean:
	rm -f example-tests framework-tests
