# Polyglot hello

Minimal Go host that:

1. Runs `ExplainReport` on the Python and Node example targets (detection score, adapter, Call honesty).
2. Calls Python `calculate` and Node `report` through one Hub.

## Requirements

- Go 1.23+
- `python3` (or `python`) on `PATH`
- `node` on `PATH`

No Cargo/Rust required (unlike `examples/showcase`).

## Run

From the repository root:

```bash
go run ./examples/polyglot-hello
```

Or:

```bash
cd examples/polyglot-hello && go run .
```

## CLI equivalent (buyer / evaluator)

```bash
make build
./bin/centralizer explain ./examples/go-python/analytics
./bin/centralizer explain ./examples/go-node/reporter
./bin/centralizer call ./examples/go-python/analytics calculate value=21
./bin/centralizer call ./examples/go-node/reporter report value=21
./bin/centralizer doctor
```

## Honesty

`explain` prints `Call implemented: yes` only for adapters that actually support invocation (Python, Node, Go, Rust protocol speakers). Detect-only languages (Lua, JVM, …) report `Call implemented: no (detect-only)` even if discovery scores confidently.
