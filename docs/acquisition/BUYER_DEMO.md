# Buyer Demo — Centralizer

**Target:** fresh machine → clone → build → explain → call → verify (≈10–15 minutes where realistic).  
**Date:** 2026-09-21  
**Contact:** [@theworker02](https://github.com/theworker02)

## Exact commands

```bash
git clone https://github.com/theworker02/centralizer.git && cd centralizer
go test ./pkg/explain ./pkg/diagnostics ./pkg/adapter ./pkg/centralizer
go test ./...
make build
./bin/centralizer version
./bin/centralizer doctor
./bin/centralizer adapters
```

### Evaluation-oriented explain (preferred buyer path)

```bash
./bin/centralizer explain ./examples/go-python/analytics
./bin/centralizer explain ./examples/go-node/reporter
```

Expect text that includes:

- detection score and chosen adapter (`python` / `node`)
- `Call implemented: yes`
- selected bridge strategy / transport
- next-step commands (`connect` / `call`)

JSON form:

```bash
./bin/centralizer --json explain ./examples/go-python/analytics
```

### Live Call (requires Python and Node on PATH)

```bash
./bin/centralizer call ./examples/go-python/analytics calculate value=21
./bin/centralizer call ./examples/go-node/reporter report value=21
go run ./examples/polyglot-hello
```

### Detect-only honesty check

```bash
# Create a tiny Lua file, or use any .lua path
echo "print('hi')" > /tmp/cz-hello.lua
./bin/centralizer explain /tmp/cz-hello.lua
```

Expect `Call implemented: no (detect-only)` and next steps that do **not** claim Call works for Lua.

## Expected results

- Commands exit 0 (or documented skip when optional runtimes are missing).
- No secrets required for the minimal path.
- `doctor` lists call-capable vs detect-only adapters.
- See [TEST_EVIDENCE.md](./TEST_EVIDENCE.md) for recorded exit codes from verification runs.
- Longer first-hour path: [../EVALUATOR_GUIDE.md](../EVALUATOR_GUIDE.md).

## Out of scope for minimal demo

- Live production cloud credentials
- Paid API quotas
- Claiming Call for detect-only adapters (Lua, JVM, C, …)
