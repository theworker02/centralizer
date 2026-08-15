# Roadmap

This is a plan, not a claim of completed work. A feature is done when code, tests, docs, errors, CI, and an example all exist.

## Phase 1 — current (v0.1)

- Go core, CIR, protocol 1.0
- Discovery with confidence scores
- stdio transport
- Go, Python, and Node adapters (invocation)
- Rust adapter for protocol-speaking binaries
- CLI and optional daemon skeleton

## Phase 2

- C and C++ process/native ABI paths that tests can run
- WASM invocation (not detect-only)
- Unix domain sockets as a live default transport
- Windows named pipes as a full server/client session (dial foundation exists)
- Supervisor recovery against real process crashes in CI on all three OSes

## Phase 3

- JVM and .NET adapters with a documented, tested call path
- Ruby and PHP stdio shims
- Planner scoring refinements from measured latency
- Broader generated-shim coverage

## Phase 4

- Shared memory (currently experimental and disabled)
- `centralizerd` service registry used by multiple host processes
- First-class streaming in language shims (Python / Node landed; Go native still planned)
- Advanced object handles (methods, properties; expiration and DropBridge landed)

## Phase 5

- Swift, Dart, Lua, Zig invocation
- Ecosystem SDKs
- Process-based third-party adapters (not Go plugins)

Do not treat later phases as implemented because files exist.
