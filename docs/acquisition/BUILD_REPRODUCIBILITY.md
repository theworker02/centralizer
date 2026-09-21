# Build Reproducibility — Centralizer

**Date:** 2026-09-21

## Fresh machine path

```
git clone https://github.com/theworker02/centralizer.git && cd centralizer
go test ./...
make build
./bin/centralizer version || ./bin/centralizer --help
```

## Assumptions

- Stack: Go 1.23+; website React/Vite; multi-language adapters
- No machine-specific absolute paths should be required.
- Cloud credentials are optional unless exercising live provider features.

## Known reproducibility limits

Documented in KNOWN_LIMITATIONS.md and project-specific diligence.
