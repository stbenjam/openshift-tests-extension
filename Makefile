all: test build

verify: lint

build:
	go build ./cmd/...

test:
	go test ./...

lint:
	./hack/go-lint.sh run ./...

clean:
	rm -f example-test
