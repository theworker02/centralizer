# Centralizer — Project-Specific Diligence

**Date:** 2026-09-21

## Supported runtimes / languages

- **Call path (Tier 1):** Go (native/process), Python, Node.js, Rust (as Protocol 1.x speaker).
- **Detect-only foundation:** C, C++, WASM, JVM, .NET, Ruby, PHP, Swift, Dart, Lua, Zig — Connect returns not-implemented per docs.

## IPC / process architecture

- Transports: stdio (NDJSON), TCP/Unix (length-prefixed JSON), Windows named pipe (experimental), SHM experimental/disabled.
- Hub → Service → supervisor → Bridge; `centralizerd` default `127.0.0.1:4780`.

## Runtime dependencies

- Go 1.23+; host interpreters/runtimes on PATH for adapters.
- `gopkg.in/yaml.v3`.

## Protocol ownership

- Centralizer Protocol 1.x is project-defined (`PROTOCOL.md`, `protocol/v1.md`, `pkg/cir`).
- Not a standards-body protocol.
- Historical Apache-tagged releases remain Apache for those artifacts — **REQUIRES_LEGAL_REVIEW** for exclusivity of protocol text.

## Performance benchmarks

- `BENCHMARKS.md` forbids publishing latency without full environment fields.
- Built-in benches are microbenches / planner scores — **not** cross-language call latency claims.

## Interoperability limitations

- Many languages are detect-only.
- Experimental transports disabled by default.

## Integration surface

- CLI + daemon + adapters + optional website docs.
- CIR schema and examples under `examples/`, `schema/`.
