BINARY := bb
PKG    := ./cmd/bb
VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install test cover lint vet vuln clean snapshot ci-local

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) $(PKG)

install:
	go install -ldflags "$(LDFLAGS)" $(PKG)

test:
	go test -race ./...

cover:
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -func=coverage.out | tail -1

lint:
	golangci-lint run

vet:
	go vet ./...

clean:
	rm -rf $(BINARY) dist/

vuln:
	@command -v govulncheck >/dev/null || (echo "installing govulncheck..." && go install golang.org/x/vuln/cmd/govulncheck@latest)
	govulncheck ./...

# Local goreleaser dry-run (produces dist/ but does not publish).
snapshot:
	goreleaser release --snapshot --clean

# Run everything CI runs. Run before pushing to catch failures locally.
ci-local: vet build cover lint vuln
	goreleaser check
	@echo
	@echo "✓ ci-local passed"
