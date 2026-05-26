BINARY := bb
PKG    := ./cmd/bb
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install test vet clean snapshot

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

install:
	go install -ldflags "$(LDFLAGS)" $(PKG)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -rf $(BINARY) dist/

# Local goreleaser dry-run (produces dist/ but does not publish).
snapshot:
	goreleaser release --snapshot --clean
