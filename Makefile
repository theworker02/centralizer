GO ?= go
BIN := bin
VERSION := 0.1.2
LDFLAGS := -X github.com/theworker02/centralizer/internal/version.Version=$(VERSION)

.PHONY: build test race lint fuzz bench docs website brand release-snapshot clean fmt vet

build:
	@mkdir -p $(BIN)
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN)/centralizer ./cmd/centralizer
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BIN)/centralizerd ./cmd/centralizerd

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	$(GO) test ./...

race:
	$(GO) test -race ./...

lint: vet
	@command -v golangci-lint >/dev/null 2>&1 && golangci-lint run ./... || echo "golangci-lint not installed; ran go vet"

fuzz:
	$(GO) test -fuzz=FuzzDecodeWire -fuzztime=10s ./pkg/cir
	$(GO) test -fuzz=FuzzParse -fuzztime=5s ./pkg/manifest
	$(GO) test -fuzz=FuzzParseYAML -fuzztime=5s ./pkg/schema
	$(GO) test -fuzz=FuzzReadNDJSON -fuzztime=5s ./internal/protocol
	$(GO) test -fuzz=FuzzReadFrame -fuzztime=5s ./internal/protocol

bench:
	$(GO) test -bench=. -benchmem ./benchmarks/... ./pkg/cir ./internal/protocol

docs:
	@echo "See README.md and docs/ for documentation."

brand:
	node scripts/sync-brand.mjs

website: brand
	cd website && npm install && npm run build

release-snapshot:
	@command -v goreleaser >/dev/null 2>&1 && goreleaser release --snapshot --clean || echo "goreleaser not installed"

clean:
	rm -rf $(BIN) dist coverage.out
