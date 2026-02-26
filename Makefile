BINARY       := mm-channel-heatmap
MODULE       := github.com/jlandells/mm-channel-heatmap
VERSION      := $(shell cat VERSION)
EXISTING_TAG := $(shell git tag -l "$(VERSION)")
LDFLAGS      := -X 'main.Version=$(VERSION)'

.PHONY: pre-build-check build build-all test test-cover lint clean

pre-build-check:
	@if [ -n "$(EXISTING_TAG)" ]; then \
		echo "ERROR: tag $(VERSION) already exists. Bump VERSION before building."; \
		exit 1; \
	fi

build: pre-build-check
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

build-all: pre-build-check
	GOOS=darwin  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-darwin-arm64  .
	GOOS=darwin  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-darwin-amd64  .
	GOOS=linux   GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-amd64   .
	GOOS=linux   GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY)-linux-arm64   .

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY) coverage.out
	rm -rf bin/
