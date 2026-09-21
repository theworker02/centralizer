# Buyer Demo — Centralizer

**Target:** fresh machine → clone → install → run → verify (≈10–15 minutes where realistic).

## Exact commands

```bash
git clone https://github.com/theworker02/centralizer.git && cd centralizer
go test ./...
make build
./bin/centralizer version || ./bin/centralizer --help
```

## Expected results

- Commands exit 0 (or documented skip for optional live-cloud steps).
- No secrets required for the minimal path.
- See TEST_EVIDENCE.md for recorded exit codes from this program’s verification runs.

## Out of scope for minimal demo

- Live production cloud credentials
- Shipping malware / real vehicle bus hardware (OpenDashCAN)
- Paid API quotas
