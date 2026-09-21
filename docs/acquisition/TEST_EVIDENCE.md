# Test Evidence — Centralizer

**Date:** 2026-09-21

## Commands run

```
go test ./...   # exit 0
make build      # exit 0 → bin/centralizer, bin/centralizerd
./bin/centralizer version  # centralizer 0.1.2 (protocol 1.0)
```

## Results

| Check | Status | Detail |
|-------|--------|--------|
| go test ./... | VERIFIED | all packages ok |
| make build | VERIFIED | binaries produced |
| version | VERIFIED | 0.1.2 / protocol 1.0 |
