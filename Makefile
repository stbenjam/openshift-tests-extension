GO_PKG_NAME := github.com/openshift-eng/openshift-tests-extension

GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_TREE_STATE := $(shell if git diff --quiet; then echo clean; else echo dirty; fi)

LDFLAGS := -X '$(GO_PKG_NAME)/pkg/extensions.CommitFromGit=$(GIT_COMMIT)' \
           -X '$(GO_PKG_NAME)/pkg/extensions.BuildDate=$(BUILD_DATE)' \
           -X '$(GO_PKG_NAME)/pkg/extensions.GitTreeState=$(GIT_TREE_STATE)'

all: test build

verify: lint

build:
	go build -ldflags "$(LDFLAGS)" ./cmd/...

test:
	go test ./...

lint:
	./hack/go-lint.sh run ./...

clean:
	rm -f example-test
